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

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockWalletRepository) Store(wallet *model.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) Find(id uuid.UUID) (*model.Wallet, error) {
	args := m.Called(id)
	if wallet, ok := args.Get(0).(*model.Wallet); ok {
		return wallet, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockWalletRepository) Remove(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func newWallet(id, userID uuid.UUID, balance float64) *model.Wallet {
	now := time.Now()
	return &model.Wallet{
		ID:        id,
		UserID:    userID,
		Balance:   balance,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCreateWallet_Success(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDispatcher := new(MockEventDispatcher)

	userID := uuid.New()
	walletID := uuid.New()

	walletRepo.On("NextID").Return(walletID, nil)
	walletRepo.On("Store", mock.MatchedBy(func(wallet *model.Wallet) bool {
		return wallet.ID == walletID &&
			wallet.UserID == userID &&
			wallet.Balance == defaultBalance &&
			!wallet.CreatedAt.IsZero() &&
			wallet.UpdatedAt.Equal(wallet.CreatedAt)
	})).Return(nil)

	eventDispatcher.On("Dispatch", mock.MatchedBy(func(e model.WalletCreated) bool {
		return e.WalletID == walletID && e.UserID == userID && e.Balance == defaultBalance
	})).Return(nil)

	svc := NewWalletService(walletRepo, eventDispatcher)

	id, err := svc.CreateWallet(userID)

	assert.NoError(t, err)
	assert.Equal(t, walletID, id)
	walletRepo.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}

func TestCreateWallet_RepoError(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	walletID := uuid.New()

	walletRepo.On("NextID").Return(walletID, nil)
	walletRepo.On("Store", mock.Anything).Return(errors.New("db down"))

	svc := NewWalletService(walletRepo, eventDisp)

	_, err := svc.CreateWallet(userID)
	assert.Error(t, err)
	walletRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestCreateWallet_EventDispatchError(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	walletID := uuid.New()

	walletRepo.On("NextID").Return(walletID, nil)
	walletRepo.On("Store", mock.Anything).Return(nil)
	eventDisp.On("Dispatch", mock.Anything).Return(errors.New("kafka unreachable"))

	svc := NewWalletService(walletRepo, eventDisp)

	_, err := svc.CreateWallet(userID)
	assert.Error(t, err)
	walletRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveWallet_Success(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	walletID := uuid.New()
	userID := uuid.New()
	wallet := newWallet(walletID, userID, 100.0)

	walletRepo.On("Find", walletID).Return(wallet, nil)
	walletRepo.On("Store", mock.MatchedBy(func(w *model.Wallet) bool {
		return w.ID == walletID && w.DeletedAt != nil
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.WalletRemoved) bool {
		return e.WalletID == walletID
	})).Return(nil)

	svc := NewWalletService(walletRepo, eventDisp)

	err := svc.RemoveWallet(walletID)
	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveWallet_NotFound_Idempotent(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	walletID := uuid.New()
	walletRepo.On("Find", walletID).Return(nil, ErrWalletNotFound)

	svc := NewWalletService(walletRepo, eventDisp)

	err := svc.RemoveWallet(walletID)
	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
}

func TestRemoveWallet_FindError(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	walletID := uuid.New()
	walletRepo.On("Find", walletID).Return(nil, errors.New("db timeout"))

	svc := NewWalletService(walletRepo, eventDisp)

	err := svc.RemoveWallet(walletID)
	assert.Error(t, err)
}

func TestUpdateWalletBalance_Success(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	walletID := uuid.New()
	userID := uuid.New()
	oldBalance := 100.0
	newBalance := 150.0

	wallet := newWallet(walletID, userID, oldBalance)
	walletRepo.On("Find", walletID).Return(wallet, nil)
	walletRepo.On("Store", mock.MatchedBy(func(w *model.Wallet) bool {
		return w.ID == walletID && w.Balance == newBalance && w.UpdatedAt.After(w.CreatedAt)
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.WalletBalanceChanged) bool {
		return e.WalletID == walletID && e.OldBalance == oldBalance && e.NewBalance == newBalance
	})).Return(nil)

	svc := NewWalletService(walletRepo, eventDisp)

	err := svc.UpdateWalletBalance(walletID, newBalance)
	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestUpdateWalletBalance_WalletNotFound(t *testing.T) {
	walletRepo := new(MockWalletRepository)
	eventDisp := new(MockEventDispatcher)

	walletID := uuid.New()
	newBalance := 150.0

	walletRepo.On("Find", walletID).Return(nil, ErrWalletNotFound)

	svc := NewWalletService(walletRepo, eventDisp)

	err := svc.UpdateWalletBalance(walletID, newBalance)
	assert.ErrorIs(t, err, ErrWalletNotFound)
	walletRepo.AssertExpectations(t)
}
