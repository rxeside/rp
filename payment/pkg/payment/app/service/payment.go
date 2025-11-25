package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/app/data"
	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/domain/service"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, orderID uuid.UUID, amount float64) (uuid.UUID, error)
	RemovePayment(ctx context.Context, paymentID uuid.UUID) error
	SetPaymentStatus(ctx context.Context, paymentID uuid.UUID, status int) error
	FindPayment(ctx context.Context, paymentID uuid.UUID) (data.Payment, error)
}

func NewPaymentService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) PaymentService {
	return &paymentService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type paymentService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *paymentService) CreatePayment(ctx context.Context, orderID uuid.UUID, amount float64) (uuid.UUID, error) {
	var paymentID uuid.UUID

	err := s.luow.Execute(ctx, []string{paymentLockByOrder(orderID)}, func(provider RepositoryProvider) error {
		domainService := s.paymentDomainService(ctx, provider.PaymentRepository(ctx))
		id, err := domainService.CreatePayment(orderID, amount)
		if err != nil {
			return err
		}
		paymentID = id
		return nil
	})

	return paymentID, err
}

func (s *paymentService) RemovePayment(ctx context.Context, paymentID uuid.UUID) error {
	return s.luow.Execute(ctx, []string{paymentLock(paymentID)}, func(provider RepositoryProvider) error {
		return s.paymentDomainService(ctx, provider.PaymentRepository(ctx)).RemovePayment(paymentID)
	})
}

func (s *paymentService) SetPaymentStatus(ctx context.Context, paymentID uuid.UUID, status int) error {
	return s.luow.Execute(ctx, []string{paymentLock(paymentID)}, func(provider RepositoryProvider) error {
		return s.paymentDomainService(ctx, provider.PaymentRepository(ctx)).SetStatus(paymentID, model.PaymentStatus(status))
	})
}

func (s *paymentService) FindPayment(ctx context.Context, paymentID uuid.UUID) (data.Payment, error) {
	var payment data.Payment
	err := s.luow.Execute(ctx, []string{paymentLock(paymentID)}, func(provider RepositoryProvider) error {
		domainPayment, err := provider.PaymentRepository(ctx).Find(paymentID)
		if err != nil {
			return err
		}
		payment = data.Payment{
			ID:       domainPayment.ID,
			WalletID: domainPayment.WalletID,
			OrderID:  domainPayment.OrderID,
			Amount:   domainPayment.Amount,
			Status:   data.PaymentStatus(domainPayment.Status),
		}
		return nil
	})
	return payment, err
}

func (s *paymentService) paymentDomainService(ctx context.Context, repository model.PaymentRepository) service.Payment {
	return service.NewPaymentService(repository, s.domainEventDispatcher(ctx))
}

func (s *paymentService) domainEventDispatcher(ctx context.Context) commonevent.Dispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const basePaymentLock = "payment_"

func paymentLock(id uuid.UUID) string {
	return basePaymentLock + id.String()
}

func paymentLockByOrder(orderID uuid.UUID) string {
	return basePaymentLock + "order_" + orderID.String()
}
