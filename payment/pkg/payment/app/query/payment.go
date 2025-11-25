package query

import (
	"context"

	"github.com/google/uuid"

	"payment/pkg/payment/app/data"
)

type PaymentQueryService interface {
	FindPayment(ctx context.Context, paymentID uuid.UUID) (*data.Payment, error)
}
