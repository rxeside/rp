package transport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order/api/server/orderinternalapi"
	appdata "order/pkg/order/app/data"
	appquery "order/pkg/order/app/query"
	appservice "order/pkg/order/app/service"
)

func NewOrderInternalAPI(
	orderQueryService appquery.OrderQueryService,
	orderService appservice.OrderService,
) orderinternalapi.OrderInternalAPIServer {
	return &orderInternalAPI{
		orderQueryService: orderQueryService,
		orderService:      orderService,
	}
}

type orderInternalAPI struct {
	orderQueryService appquery.OrderQueryService
	orderService      appservice.OrderService

	orderinternalapi.UnimplementedOrderInternalAPIServer
}

func (o orderInternalAPI) StoreOrder(ctx context.Context, request *orderinternalapi.StoreOrderRequest) (*orderinternalapi.StoreOrderResponse, error) {
	var (
		orderID uuid.UUID
		err     error
	)
	if request.OrderID != "" {
		orderID, err = uuid.Parse(request.OrderID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.OrderID)
		}
	}

	customerID, err := uuid.Parse(request.CustomerID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.CustomerID)
	}

	items := make([]appdata.OrderItem, len(request.Items))
	for i, item := range request.Items {
		orderItemID, err := uuid.Parse(item.OrderID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", item.OrderID)
		}
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", item.ProductID)
		}
		items[i] = appdata.OrderItem{
			OrderID:    orderItemID,
			ProductID:  productID,
			Count:      int(item.Count),
			TotalPrice: item.TotalPrice,
		}
	}

	orderID, err = o.orderService.StoreOrder(ctx, appdata.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     appdata.OrderStatus(request.Status),
		Items:      items,
	})
	if err != nil {
		return nil, err
	}

	return &orderinternalapi.StoreOrderResponse{
		OrderID: orderID.String(),
	}, nil
}

func (o orderInternalAPI) FindOrder(ctx context.Context, request *orderinternalapi.FindOrderRequest) (*orderinternalapi.FindOrderResponse, error) {
	orderID, err := uuid.Parse(request.OrderID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.OrderID)
	}
	order, err := o.orderQueryService.FindUser(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, status.Errorf(codes.NotFound, "order %q not found", request.OrderID)
	}

	items := make([]*orderinternalapi.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &orderinternalapi.OrderItem{
			OrderID:    item.OrderID.String(),
			ProductID:  item.ProductID.String(),
			Count:      int32(item.Count), // #nosec G115
			TotalPrice: item.TotalPrice,
		}
	}

	response := &orderinternalapi.FindOrderResponse{
		OrderID:    orderID.String(),
		Status:     orderinternalapi.OrderStatus(order.Status), // nolint:gosec
		CustomerID: order.CustomerID.String(),
		Items:      items,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  order.UpdatedAt.Format(time.RFC3339),
	}
	if order.DeletedAt != nil {
		deletedAtStr := order.DeletedAt.Format(time.RFC3339)
		response.DeletedAt = &deletedAtStr
	}

	return response, nil
}
