package data

import "github.com/google/uuid"

type UserStatus int

const (
	Blocked UserStatus = iota
	Active
	Deleted
)

type User struct {
	UserID   uuid.UUID
	Status   UserStatus
	Login    string
	Email    *string
	Telegram *string
}
