package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "order/pkg/common/event"
	"order/pkg/order/domain/model"
)

var (
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

type OrderService interface {
	CreateOrder(customerID uuid.UUID) (uuid.UUID, error)
	RemoveOrder(orderID uuid.UUID) error
	SetStatus(orderID uuid.UUID, status model.OrderStatus) error
	AddItem(orderID, productID uuid.UUID, price float64) error
	RemoveItem(orderID, itemID uuid.UUID) error
}

func NewOrderService(repo model.OrderRepository, dispatcher commonevent.Dispatcher) OrderService {
	return &orderService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type orderService struct {
	repo       model.OrderRepository
	dispatcher commonevent.Dispatcher
}

func (o orderService) CreateOrder(customerID uuid.UUID) (uuid.UUID, error) {
	orderID, err := o.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = o.repo.Store(&model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.Open,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, o.dispatcher.Dispatch(model.OrderCreated{
		OrderID:    orderID,
		CustomerID: customerID,
	})
}

func (o orderService) RemoveOrder(orderID uuid.UUID) error {
	order, err := o.repo.Find(orderID)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	order.DeletedAt = &now
	order.UpdatedAt = now

	if err = o.repo.Store(order); err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderRemoved{
		OrderID: orderID,
	})
}

func (o orderService) SetStatus(orderID uuid.UUID, status model.OrderStatus) error {
	order, err := o.repo.Find(orderID)
	if err != nil {
		return err
	}

	oldStatus := order.Status

	if !o.isValidStatusTransition(order.Status, status) {
		return ErrInvalidOrderStatus
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err = o.repo.Store(order); err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderStatusChanged{
		OrderID: orderID,
		From:    oldStatus,
		To:      status,
	})
}

func (o orderService) AddItem(orderID, productID uuid.UUID, price float64) error {
	order, err := o.repo.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.Open {
		return ErrInvalidOrderStatus
	}

	order.Items = append(order.Items, model.OrderItem{
		OrderID:    orderID,
		ProductID:  productID,
		TotalPrice: price,
	})

	err = o.repo.Store(order)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderItemsChanged{
		OrderID:    orderID,
		AddedItems: []uuid.UUID{productID},
	})
}

func (o orderService) RemoveItem(orderID, itemID uuid.UUID) error {
	order, err := o.repo.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.Open {
		return ErrInvalidOrderStatus
	}

	found := false
	var newItems []model.OrderItem
	for _, item := range order.Items {
		if item.ProductID != itemID {
			newItems = append(newItems, item)
		} else {
			found = true
		}
	}

	if !found {
		return nil
	}

	order.Items = newItems
	order.UpdatedAt = time.Now()

	if err = o.repo.Store(order); err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderItemsChanged{
		OrderID:      orderID,
		RemovedItems: []uuid.UUID{itemID},
	})
}

func (o orderService) isValidStatusTransition(from, to model.OrderStatus) bool {
	switch from {
	case model.Open:
		return to == model.Open || to == model.Pending || to == model.Paid || to == model.Cancelled
	case model.Pending:
		return to == model.Paid || to == model.Cancelled
	case model.Paid, model.Cancelled:
		return false
	default:
		return false
	}
}
