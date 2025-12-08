package model

import "github.com/google/uuid"

type ProductCreated struct {
	ProductID uuid.UUID
	Name      string
	Price     float64
}

func (e ProductCreated) Type() string {
	return "ProductCreated"
}

type ProductUpdated struct {
	ProductID uuid.UUID
}

func (e ProductUpdated) Type() string {
	return "ProductUpdated"
}

type ProductRemoved struct {
	ProductID uuid.UUID
}

func (e ProductRemoved) Type() string {
	return "ProductRemoved"
}
