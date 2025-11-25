package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"payment/pkg/payment/app/service"
	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) WalletRepository(ctx context.Context) model.WalletRepository {
	return repository.NewWalletRepository(ctx, r.client)
}

func (r *repositoryProvider) PaymentRepository(ctx context.Context) model.PaymentRepository {
	return repository.NewPaymentRepository(ctx, r.client)
}
