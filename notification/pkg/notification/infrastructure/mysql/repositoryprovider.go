package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"notification/pkg/notification/app/service"
	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) NotificationRepository(ctx context.Context) model.NotificationRepository {
	return repository.NewNotificationRepository(ctx, r.client)
}

func (r *repositoryProvider) UserRepository(ctx context.Context) model.UserRepository {
	return repository.NewUserRepository(ctx, r.client)
}
