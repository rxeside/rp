package model

import "github.com/google/uuid"

type WalletCreated struct {
	WalletID uuid.UUID
	UserID   uuid.UUID
	Balance  float64
}

func (e WalletCreated) Type() string {
	return "WalletCreated"
}

type WalletBalanceChanged struct {
	WalletID   uuid.UUID
	OldBalance float64
	NewBalance float64
}

func (e WalletBalanceChanged) Type() string {
	return "WalletBalanceChanged"
}

type WalletRemoved struct {
	WalletID uuid.UUID
}

func (e WalletRemoved) Type() string {
	return "WalletRemoved"
}

type PaymentCreated struct {
	PaymentID uuid.UUID
	WalletID  uuid.UUID
	OrderID   uuid.UUID
	Amount    float64
}

func (e PaymentCreated) Type() string {
	return "PaymentCreated"
}

type PaymentStatusChanged struct {
	PaymentID uuid.UUID
	From      PaymentStatus
	To        PaymentStatus
}

func (e PaymentStatusChanged) Type() string {
	return "PaymentStatusChanged"
}

type PaymentRemoved struct {
	PaymentID uuid.UUID
}

func (e PaymentRemoved) Type() string {
	return "PaymentRemoved"
}
