package transport

import (
	"payment/api/server/paymentinternal"
	"payment/pkg/payment/app/query"
	"payment/pkg/payment/app/service"
)

func NewPaymentInternalAPI(
	paymentQueryService query.PaymentQueryService,
	paymentService service.PaymentService,
	walletQueryService query.WalletQueryService,
	walletService service.WalletService,
) paymentinternal.PaymentInternalAPIServer {
	return &paymentInternalAPI{
		paymentQueryService: paymentQueryService,
		paymentService:      paymentService,
		walletQueryService:  walletQueryService,
		walletService:       walletService,
	}
}

type paymentInternalAPI struct {
	paymentQueryService query.PaymentQueryService
	paymentService      service.PaymentService
	walletQueryService  query.WalletQueryService
	walletService       service.WalletService

	paymentinternal.UnsafePaymentInternalAPIServer
}
