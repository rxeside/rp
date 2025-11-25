package query

import (
	"context"

	"github.com/google/uuid"

	"payment/pkg/payment/app/data"
)

type WalletQueryService interface {
	FindWallet(ctx context.Context, walletID uuid.UUID) (*data.Wallet, error)
}
