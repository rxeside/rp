package model

import "github.com/google/uuid"

type NotificationCreated struct {
	NotificationID uuid.UUID
	UserID         uuid.UUID
	OrderID        uuid.UUID
	Status         string
	Message        string
}

func (e NotificationCreated) EventType() string {
	return "NotificationCreated"
}

type NotificationStatusChanged struct {
	NotificationID uuid.UUID
	From           string
	To             string
}

func (e NotificationStatusChanged) EventType() string {
	return "NotificationStatusChanged"
}

type NotificationRemoved struct {
	NotificationID uuid.UUID
}

func (e NotificationRemoved) EventType() string {
	return "NotificationRemoved"
}
