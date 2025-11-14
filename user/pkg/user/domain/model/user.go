package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        uuid.UUID
	Login     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type UserRepository interface {
	NextID() (uuid.UUID, error)
	Store(user *User) error
	Find(id uuid.UUID) (*User, error)
	Remove(id uuid.UUID) error
}
