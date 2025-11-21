package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "payment/pkg/common/event"
	"payment/pkg/payment/domain/model"
)

const defaultBalance = 100000.0

var (
	ErrInvalidWalletBalance = errors.New("invalid wallet balance")
	ErrWalletNotFound       = errors.New("wallet not found")
)

type Wallet interface {
	CreateWallet(userID uuid.UUID) (uuid.UUID, error)
	RemoveWallet(walletID uuid.UUID) error
	UpdateWalletBalance(walletID uuid.UUID, newBalance float64) error
}

func NewWalletService(repo model.WalletRepository, dispatcher commonevent.Dispatcher) Wallet {
	return &walletService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type walletService struct {
	repo       model.WalletRepository
	dispatcher commonevent.Dispatcher
}

func (w walletService) CreateWallet(userID uuid.UUID) (uuid.UUID, error) {
	walletID, err := w.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	initialBalance := defaultBalance
	currentTime := time.Now()
	err = w.repo.Store(&model.Wallet{
		ID:        walletID,
		UserID:    userID,
		Balance:   initialBalance,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return walletID, w.dispatcher.Dispatch(model.WalletCreated{
		WalletID: walletID,
		UserID:   userID,
		Balance:  initialBalance,
	})
}

func (w walletService) RemoveWallet(walletID uuid.UUID) error {
	wallet, err := w.repo.Find(walletID)
	if err != nil {
		if errors.Is(err, ErrWalletNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	wallet.DeletedAt = &now
	wallet.UpdatedAt = now

	if err = w.repo.Store(wallet); err != nil {
		return err
	}

	return w.dispatcher.Dispatch(model.WalletRemoved{
		WalletID: walletID,
	})
}

func (w walletService) UpdateWalletBalance(walletID uuid.UUID, newBalance float64) error {
	wallet, err := w.repo.Find(walletID)
	if err != nil {
		return err
	}

	oldBalance := wallet.Balance

	if newBalance < 0 {
		return ErrInvalidWalletBalance
	}

	wallet.Balance = newBalance
	wallet.UpdatedAt = time.Now()

	if err = w.repo.Store(wallet); err != nil {
		return err
	}

	return w.dispatcher.Dispatch(model.WalletBalanceChanged{
		WalletID:   walletID,
		OldBalance: oldBalance,
		NewBalance: newBalance,
	})
}
