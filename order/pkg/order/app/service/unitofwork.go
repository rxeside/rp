package service

import (
	"context"

	"order/pkg/order/domain/model"
)

type RepositoryProvider interface {
	OrderRepository(ctx context.Context) model.OrderRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
