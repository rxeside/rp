package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/domain/model"
)

var (
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
)

type Payment interface {
	CreatePayment(orderID uuid.UUID, amount float64) (uuid.UUID, error)
	RemovePayment(paymentID uuid.UUID) error
	SetStatus(paymentID uuid.UUID, status model.PaymentStatus) error
}

func NewPaymentService(repo model.PaymentRepository, dispatcher commonevent.Dispatcher) Payment {
	return &paymentService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type paymentService struct {
	repo       model.PaymentRepository
	dispatcher commonevent.Dispatcher
}

func (p paymentService) CreatePayment(orderID uuid.UUID, amount float64) (uuid.UUID, error) {
	paymentID, err := p.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = p.repo.Store(&model.Payment{
		ID:        paymentID,
		OrderID:   orderID,
		Amount:    amount,
		Status:    model.Pending,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return paymentID, p.dispatcher.Dispatch(model.PaymentCreated{
		PaymentID: paymentID,
		OrderID:   orderID,
		Amount:    amount,
	})
}

func (p paymentService) RemovePayment(paymentID uuid.UUID) error {
	payment, err := p.repo.Find(paymentID)
	if err != nil {
		if errors.Is(err, model.ErrPaymentNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	payment.DeletedAt = &now
	payment.UpdatedAt = now

	if err = p.repo.Store(payment); err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.PaymentRemoved{
		PaymentID: paymentID,
	})
}

func (p paymentService) SetStatus(paymentID uuid.UUID, status model.PaymentStatus) error {
	payment, err := p.repo.Find(paymentID)
	if err != nil {
		return err
	}

	oldStatus := payment.Status

	if !p.isValidStatusTransition(payment.Status, status) {
		return ErrInvalidPaymentStatus
	}

	payment.Status = status
	payment.UpdatedAt = time.Now()

	if err = p.repo.Store(payment); err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.PaymentStatusChanged{
		PaymentID: paymentID,
		From:      oldStatus,
		To:        status,
	})
}

func (p paymentService) isValidStatusTransition(from, to model.PaymentStatus) bool {
	switch from {
	case model.Pending:
		return to == model.Processing || to == model.Cancelled
	case model.Processing:
		return to == model.Succeeded || to == model.Failed
	case model.Succeeded, model.Failed, model.Cancelled:
		return false
	default:
		return false
	}
}
