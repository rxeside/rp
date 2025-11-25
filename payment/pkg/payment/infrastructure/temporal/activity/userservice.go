package activity

import (
	"context"

	"github.com/google/uuid"

	appdata "payment/pkg/payment/app/data"
	"payment/pkg/payment/app/service"
)

func NewPaymentServiceActivities(paymentService service.PaymentService) *PaymentServiceActivities {
	return &PaymentServiceActivities{paymentService: paymentService}
}

type PaymentServiceActivities struct {
	paymentService service.PaymentService
}

func (a *PaymentServiceActivities) FindUser(ctx context.Context, userID uuid.UUID) (appdata.Payment, error) {
	return a.paymentService.FindPayment(ctx, userID)
}
