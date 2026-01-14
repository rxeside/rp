package activity

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"payment/pkg/payment/app/service"
)

func NewWalletServiceActivities(walletService service.WalletService) *WalletServiceActivities {
	return &WalletServiceActivities{
		walletService: walletService,
	}
}

type WalletServiceActivities struct {
	walletService service.WalletService
}

func (a *WalletServiceActivities) CreateWallet(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	fmt.Println("CreateWallet userID = ", userID)
	return a.walletService.CreateWallet(ctx, userID)
}

func (a *WalletServiceActivities) ChargeWallet(ctx context.Context, userIDStr string, amount float64) error {
	uid, _ := uuid.Parse(userIDStr)
	return a.walletService.UpdateWalletBalance(ctx, uid, -amount)
}

func (a *WalletServiceActivities) RefundWallet(ctx context.Context, userIDStr string, amount float64) error {
	uid, _ := uuid.Parse(userIDStr)
	return a.walletService.UpdateWalletBalance(ctx, uid, amount)
}
