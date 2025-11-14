package model

import "github.com/google/uuid"

type OrderCreated struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
}

func (e OrderCreated) EventType() string {
	return "OrderCreated"
}

type OrderItemsChanged struct {
	OrderID      uuid.UUID
	AddedItems   []uuid.UUID
	RemovedItems []uuid.UUID
}

func (e OrderItemsChanged) EventType() string {
	return "OrderItemsChanged"
}

type OrderRemoved struct {
	OrderID uuid.UUID
}

func (e OrderRemoved) EventType() string {
	return "OrderRemoved"
}

type OrderStatusChanged struct {
	OrderID uuid.UUID
	From    OrderStatus
	To      OrderStatus
}

func (e OrderStatusChanged) EventType() string {
	return "OrderStatusChanged"
}
