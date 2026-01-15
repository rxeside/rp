package service

import (
	"context"
	"order/pkg/order/infrastructure/temporal/workflows"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client" // Не забываем импорт

	commonevent "order/pkg/common/event"
	appdata "order/pkg/order/app/data"
	"order/pkg/order/domain/model"
	"order/pkg/order/domain/service"
)

type OrderService interface {
	StoreOrder(ctx context.Context, order appdata.Order) (uuid.UUID, error)
	SetOrderStatus(ctx context.Context, orderID uuid.UUID, status int) error
	FindOrder(ctx context.Context, orderID uuid.UUID) (appdata.Order, error)
}

func NewOrderService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
	temporalClient client.Client, // Добавляем в параметры
) OrderService {
	return &orderService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
		temporalClient:  temporalClient, // Сохраняем клиента
	}
}

type orderService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
	temporalClient  client.Client // Добавляем поле в структуру
}

func (s *orderService) StoreOrder(ctx context.Context, order appdata.Order) (uuid.UUID, error) {
	orderID := order.ID
	err := s.luow.Execute(ctx, []string{orderLock(orderID)}, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider.OrderRepository(ctx))
		if order.ID == uuid.Nil {
			oID, err := domainService.CreateOrder(order.CustomerID)
			if err != nil {
				return err
			}
			orderID = oID
		}

		err := domainService.SetStatus(orderID, model.OrderStatus(order.Status))
		if err != nil {
			return err
		}

		for _, item := range order.Items {
			err := domainService.AddItem(orderID, item.ProductID, item.TotalPrice)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err == nil && order.Status == appdata.Open {
		items := make([]workflows.OrderItemParam, len(order.Items))
		var total float64
		for i, it := range order.Items {
			items[i] = workflows.OrderItemParam{ProductID: it.ProductID.String(), Quantity: it.Count}
			total += it.TotalPrice
		}

		_, sagaErr := s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
			ID:        "order-saga-" + orderID.String(),
			TaskQueue: "order_task_queue",
		}, workflows.CreateOrderSaga, workflows.OrderSagaParams{
			OrderID:    orderID.String(),
			UserID:     order.CustomerID.String(),
			Items:      items,
			TotalPrice: total,
		})

		if sagaErr != nil {
			return orderID, sagaErr
		}
	}

	return orderID, err
}

func (s *orderService) SetOrderStatus(ctx context.Context, orderID uuid.UUID, status int) error {
	return s.luow.Execute(ctx, []string{orderLock(orderID)}, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider.OrderRepository(ctx)).SetStatus(orderID, model.OrderStatus(status))
	})
}

func (s *orderService) FindOrder(ctx context.Context, orderID uuid.UUID) (appdata.Order, error) {
	var order appdata.Order
	err := s.luow.Execute(ctx, []string{orderLock(orderID)}, func(provider RepositoryProvider) error {
		domainOrder, err := provider.OrderRepository(ctx).Find(orderID)
		if err != nil {
			return err
		}
		order = appdata.Order{
			ID:         domainOrder.ID,
			CustomerID: domainOrder.CustomerID,
			Status:     appdata.OrderStatus(domainOrder.Status),
			Items:      make([]appdata.OrderItem, len(domainOrder.Items)),
			CreatedAt:  domainOrder.CreatedAt,
			UpdatedAt:  domainOrder.UpdatedAt,
			DeletedAt:  domainOrder.DeletedAt,
		}
		for i, item := range domainOrder.Items {
			order.Items[i] = appdata.OrderItem{
				OrderID:    item.OrderID,
				ProductID:  item.ProductID,
				Count:      item.Count,
				TotalPrice: item.TotalPrice,
			}
		}
		return nil
	})
	return order, err
}

func (s *orderService) domainService(ctx context.Context, repository model.OrderRepository) service.OrderService {
	return service.NewOrderService(repository, s.domainEventDispatcher(ctx))
}

func (s *orderService) domainEventDispatcher(ctx context.Context) commonevent.Dispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseOrderLock = "order_"

func orderLock(id uuid.UUID) string {
	return baseOrderLock + id.String()
}
