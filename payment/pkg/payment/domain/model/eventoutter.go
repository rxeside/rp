package model

import (
	"time"

	"github.com/google/uuid"
)

type UserCreated struct {
	UserID    string  `json:"user_id"`
	Status    int     `json:"status"`
	Login     string  `json:"login"`
	Email     *string `json:"email,omitempty"`
	Telegram  *string `json:"telegram,omitempty"`
	CreatedAt int64   `json:"created_at"`
}

func (u UserCreated) Type() string {
	return "user_created"
}

type UserUpdated struct {
	UserID        string `json:"user_id"`
	UpdatedFields *struct {
		Status   *int    `json:"status,omitempty"`
		Email    *string `json:"email,omitempty"`
		Telegram *string `json:"telegram,omitempty"`
	} `json:"updated_fields,omitempty"`
	RemovedFields *struct {
		Email    *bool `json:"email,omitempty"`
		Telegram *bool `json:"telegram,omitempty"`
	} `json:"removed_fields,omitempty"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

func (u UserUpdated) Type() string {
	return "user_updated"
}

type UserDeleted struct {
	UserID    string `json:"user_id"`
	Status    int    `json:"status"`
	DeletedAt int64  `json:"deleted_at"`
	Hard      bool   `json:"hard"`
}

func (u UserDeleted) Type() string {
	return "user_deleted"
}

type UserStatus int

const (
	Blocked UserStatus = iota
	Active
	Deleted
)

type User struct {
	UserID    uuid.UUID
	Status    UserStatus
	Login     string
	Email     *string
	Telegram  *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
