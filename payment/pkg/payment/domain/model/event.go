package model

import "github.com/google/uuid"

type PaymentCreated struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	Amount    float64
}

func (e PaymentCreated) EventType() string {
	return "PaymentCreated"
}

type PaymentStatusChanged struct {
	PaymentID uuid.UUID
	From      PaymentStatus
	To        PaymentStatus
}

func (e PaymentStatusChanged) EventType() string {
	return "PaymentStatusChanged"
}

type PaymentRemoved struct {
	PaymentID uuid.UUID
}

func (e PaymentRemoved) EventType() string {
	return "PaymentRemoved"
}
