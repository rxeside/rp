package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"payment/pkg/payment/domain/model"
)

const (
	testAmount = 99.99
)

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPaymentRepository) Store(payment *model.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Find(id uuid.UUID) (*model.Payment, error) {
	args := m.Called(id)
	if payment, ok := args.Get(0).(*model.Payment); ok {
		return payment, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) Remove(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func newPendingPayment(id, orderID uuid.UUID) *model.Payment {
	now := time.Now()
	return &model.Payment{
		ID:        id,
		OrderID:   orderID,
		Amount:    testAmount,
		Status:    model.Pending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newSucceededPayment(id, orderID uuid.UUID) *model.Payment {
	payment := newPendingPayment(id, orderID)
	payment.Status = model.Succeeded
	return payment
}

func TestCreatePayment_Success(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDispatcher := new(MockEventDispatcher)

	orderID := uuid.New()
	paymentID := uuid.New()

	paymentRepo.On("NextID").Return(paymentID, nil)
	paymentRepo.On("Store", mock.MatchedBy(func(payment *model.Payment) bool {
		return payment.ID == paymentID &&
			payment.OrderID == orderID &&
			payment.Amount == testAmount &&
			payment.Status == model.Pending &&
			!payment.CreatedAt.IsZero() &&
			payment.UpdatedAt.Equal(payment.CreatedAt)
	})).Return(nil)

	eventDispatcher.On("Dispatch", mock.MatchedBy(func(e model.PaymentCreated) bool {
		return e.PaymentID == paymentID && e.OrderID == orderID && e.Amount == testAmount
	})).Return(nil)

	svc := NewPaymentService(paymentRepo, eventDispatcher)

	id, err := svc.CreatePayment(orderID, testAmount)

	assert.NoError(t, err)
	assert.Equal(t, paymentID, id)
	paymentRepo.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}

func TestCreatePayment_RepoError(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	paymentID := uuid.New()

	paymentRepo.On("NextID").Return(paymentID, nil)
	paymentRepo.On("Store", mock.Anything).Return(errors.New("db down"))

	svc := NewPaymentService(paymentRepo, eventDisp)

	_, err := svc.CreatePayment(orderID, testAmount)
	assert.Error(t, err)
	paymentRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestCreatePayment_EventDispatchError(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	paymentID := uuid.New()

	paymentRepo.On("NextID").Return(paymentID, nil)
	paymentRepo.On("Store", mock.Anything).Return(nil)
	eventDisp.On("Dispatch", mock.Anything).Return(errors.New("kafka unreachable"))

	svc := NewPaymentService(paymentRepo, eventDisp)

	_, err := svc.CreatePayment(orderID, testAmount)
	assert.Error(t, err)
	paymentRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemovePayment_Success(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	paymentID := uuid.New()
	orderID := uuid.New()
	payment := newPendingPayment(paymentID, orderID)

	paymentRepo.On("Find", paymentID).Return(payment, nil)
	paymentRepo.On("Store", mock.MatchedBy(func(p *model.Payment) bool {
		return p.ID == paymentID && p.DeletedAt != nil
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.PaymentRemoved) bool {
		return e.PaymentID == paymentID
	})).Return(nil)

	svc := NewPaymentService(paymentRepo, eventDisp)

	err := svc.RemovePayment(paymentID)
	assert.NoError(t, err)
	paymentRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemovePayment_NotFound_Idempotent(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	paymentID := uuid.New()
	paymentRepo.On("Find", paymentID).Return(nil, ErrPaymentNotFound)

	svc := NewPaymentService(paymentRepo, eventDisp)

	err := svc.RemovePayment(paymentID)
	assert.NoError(t, err)
	paymentRepo.AssertExpectations(t)
}

func TestRemovePayment_FindError(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	paymentID := uuid.New()
	paymentRepo.On("Find", paymentID).Return(nil, errors.New("db timeout"))

	svc := NewPaymentService(paymentRepo, eventDisp)

	err := svc.RemovePayment(paymentID)
	assert.Error(t, err)
}

func TestSetStatus_ValidTransition_PendingToProcessing(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	paymentID := uuid.New()
	orderID := uuid.New()
	payment := newPendingPayment(paymentID, orderID)

	paymentRepo.On("Find", paymentID).Return(payment, nil)
	paymentRepo.On("Store", mock.MatchedBy(func(p *model.Payment) bool {
		return p.Status == model.Processing
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.PaymentStatusChanged) bool {
		return e.PaymentID == paymentID && e.From == model.Pending && e.To == model.Processing
	})).Return(nil)

	svc := NewPaymentService(paymentRepo, eventDisp)

	err := svc.SetStatus(paymentID, model.Processing)
	assert.NoError(t, err)
}

func TestSetStatus_InvalidTransition_SucceededToPending(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	paymentID := uuid.New()
	orderID := uuid.New()
	payment := newSucceededPayment(paymentID, orderID)

	paymentRepo.On("Find", paymentID).Return(payment, nil)

	svc := NewPaymentService(paymentRepo, eventDisp)

	err := svc.SetStatus(paymentID, model.Pending)
	assert.ErrorIs(t, err, ErrInvalidPaymentStatus)
	paymentRepo.AssertExpectations(t)
}

func TestSetStatus_PaymentNotFound(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	eventDisp := new(MockEventDispatcher)

	paymentID := uuid.New()
	paymentRepo.On("Find", paymentID).Return(nil, ErrPaymentNotFound)

	svc := NewPaymentService(paymentRepo, eventDisp)

	err := svc.SetStatus(paymentID, model.Cancelled)
	assert.ErrorIs(t, err, ErrPaymentNotFound)
}

func TestIsValidStatusTransition_Matrix(t *testing.T) {
	svc := &paymentService{}

	tests := []struct {
		from  model.PaymentStatus
		to    model.PaymentStatus
		valid bool
		desc  string
	}{
		{model.Pending, model.Processing, true, "Pending → Processing"},
		{model.Pending, model.Cancelled, true, "Pending → Cancelled"},
		{model.Pending, model.Pending, false, "Pending → Pending (no-op)"},
		{model.Pending, model.Succeeded, false, "Pending → Succeeded (invalid)"},

		{model.Processing, model.Succeeded, true, "Processing → Succeeded"},
		{model.Processing, model.Failed, true, "Processing → Failed"},
		{model.Processing, model.Pending, false, "Processing → Pending"},
		{model.Processing, model.Processing, false, "Processing → Processing"},

		{model.Succeeded, model.Cancelled, false, "Succeeded → Cancelled"},
		{model.Succeeded, model.Pending, false, "Succeeded → Pending"},
		{model.Failed, model.Succeeded, false, "Failed → Succeeded"},
		{model.Cancelled, model.Succeeded, false, "Cancelled → Succeeded"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.valid, svc.isValidStatusTransition(tt.from, tt.to))
		})
	}
}
