package service

import (
	"context"

	"payment/pkg/payment/domain/model"
)

type RepositoryProvider interface {
	WalletRepository(ctx context.Context) model.WalletRepository
	PaymentRepository(ctx context.Context) model.PaymentRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
