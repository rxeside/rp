package activity

import (
	"context"

	"github.com/google/uuid"

	"payment/pkg/payment/app/service"
)

func NewActivities(
	paymentService service.PaymentService,
	walletService service.WalletService,
) *Activities {
	return &Activities{
		paymentService: paymentService,
		walletService:  walletService,
	}
}

type Activities struct {
	paymentService service.PaymentService
	walletService  service.WalletService
}

func (a *Activities) CreateWallet(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return a.walletService.CreateWallet(ctx, userID)
}
