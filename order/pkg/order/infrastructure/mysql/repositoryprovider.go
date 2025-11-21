package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"order/pkg/order/app/service"
	"order/pkg/order/domain/model"
	"order/pkg/order/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) OrderRepository(ctx context.Context) model.OrderRepository {
	return repository.NewOrderRepository(ctx, r.client)
}
