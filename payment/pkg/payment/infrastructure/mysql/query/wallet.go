package query

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"

	"payment/pkg/payment/app/data"
	"payment/pkg/payment/app/query"
)

func NewWalletQueryService(client mysql.ClientContext) query.WalletQueryService {
	return &walletQueryService{
		client: client,
	}
}

type walletQueryService struct {
	client mysql.ClientContext
}

func (w *walletQueryService) FindWallet(_ context.Context, _ uuid.UUID) (*data.Wallet, error) {
	return nil, nil
}
