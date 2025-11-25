package integrationevent

import (
	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(_ outbox.Event) (string, error) {
	return "", nil
}

type UserCreated struct {
	UserID    string  `json:"user_id"`
	Status    int     `json:"status"`
	Login     string  `json:"login"`
	Email     *string `json:"email,omitempty"`
	Telegram  *string `json:"telegram,omitempty"`
	CreatedAt int64   `json:"created_at"`
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

type UserDeleted struct {
	UserID    string `json:"user_id"`
	Status    int    `json:"status"`
	DeletedAt int64  `json:"deleted_at"`
	Hard      bool   `json:"hard"`
}
