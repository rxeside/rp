package model

import "github.com/google/uuid"

type UserCreated struct {
	UserID uuid.UUID
	Login  string
	Email  string
}

func (e UserCreated) EventType() string {
	return "UserCreated"
}

type UserUpdated struct {
	UserID uuid.UUID
}

func (e UserUpdated) EventType() string {
	return "UserUpdated"
}

type UserRemoved struct {
	UserID uuid.UUID
}

func (e UserRemoved) EventType() string {
	return "UserRemoved"
}
