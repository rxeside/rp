package query

import (
	"context"

	"github.com/google/uuid"

	"order/pkg/order/app/data"
)

type OrderQueryService interface {
	FindUser(ctx context.Context, orderID uuid.UUID) (*data.Order, error)
}
