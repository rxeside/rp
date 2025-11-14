package model

import "github.com/google/uuid"

type ProductCreated struct {
	ProductID uuid.UUID
	Name      string
	Price     float64
}

func (e ProductCreated) EventType() string {
	return "ProductCreated"
}

type ProductUpdated struct {
	ProductID uuid.UUID
}

func (e ProductUpdated) EventType() string {
	return "ProductUpdated"
}

type ProductRemoved struct {
	ProductID uuid.UUID
}

func (e ProductRemoved) EventType() string {
	return "ProductRemoved"
}
