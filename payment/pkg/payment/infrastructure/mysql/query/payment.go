package query

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"

	"payment/pkg/payment/app/data"
	"payment/pkg/payment/app/query"
)

func NewPaymentQueryService(client mysql.ClientContext) query.PaymentQueryService {
	return &paymentQueryService{
		client: client,
	}
}

type paymentQueryService struct {
	client mysql.ClientContext
}

func (p *paymentQueryService) FindPayment(_ context.Context, _ uuid.UUID) (*data.Payment, error) {
	return &data.Payment{}, nil
}
