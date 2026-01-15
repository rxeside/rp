package activity

import (
	"context"
	"order/pkg/order/app/service"

	"github.com/google/uuid"
)

type OrderActivities struct {
	orderService service.OrderService
}

func NewOrderActivities(os service.OrderService) *OrderActivities {
	return &OrderActivities{orderService: os}
}

func (a *OrderActivities) SetOrderStatusActivity(ctx context.Context, orderID string, status string) error {
	uid, _ := uuid.Parse(orderID)
	statusMap := map[string]int{"Open": 0, "Pending": 1, "Paid": 2, "Cancelled": 3}
	return a.orderService.SetOrderStatus(ctx, uid, statusMap[status])
}
