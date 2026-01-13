package service

import (
	"context"

	"notification/pkg/notification/domain/model"
)

type RepositoryProvider interface {
	NotificationRepository(ctx context.Context) model.NotificationRepository
	UserRepository(ctx context.Context) model.UserRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
