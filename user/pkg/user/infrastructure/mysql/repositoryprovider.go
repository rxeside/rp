package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"user/pkg/user/application/service"
	"user/pkg/user/domain/model"
	"user/pkg/user/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) UserRepository(ctx context.Context) model.UserRepository {
	return repository.NewUserRepository(ctx, r.client)
}
