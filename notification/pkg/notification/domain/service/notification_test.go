package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notification/pkg/notification/domain/model"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRepo) Store(n model.Notification) error {
	args := m.Called(n)
	return args.Error(0)
}

func (m *mockRepo) Find(id uuid.UUID) (*model.Notification, error) {
	args := m.Called(id)
	if notif, ok := args.Get(0).(*model.Notification); ok {
		return notif, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCreateNotification(t *testing.T) {
	repo := new(mockRepo)
	svc := NewNotificationService(repo)

	id := uuid.New()
	payload := model.NotificationPayload{
		Email:   "test@example.com",
		Message: "Hello!",
	}
	now := time.Now()

	repo.On("NextID").Return(id, nil)
	repo.On("Store", mock.MatchedBy(func(n model.Notification) bool {
		return n.ID == id &&
			n.Payload.Email == payload.Email &&
			n.Payload.Message == payload.Message &&
			n.Status == nil &&
			n.ExecutedAt == nil &&
			n.CreatedAt.After(now.Add(-time.Second)) &&
			n.UpdatedAt.After(now.Add(-time.Second))
	})).Return(nil)

	gotID, err := svc.CreateNotification(payload)
	assert.NoError(t, err)
	assert.Equal(t, id, gotID)
}

func TestMarkAsExecuted_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewNotificationService(repo)

	id := uuid.New()
	now := time.Now()
	notif := &model.Notification{
		ID:        id,
		Payload:   model.NotificationPayload{Email: "a@b.com", Message: "msg"},
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	repo.On("Find", id).Return(notif, nil)
	repo.On("Store", mock.MatchedBy(func(n model.Notification) bool {
		return n.ID == id &&
			n.Status != nil &&
			*n.Status == model.StatusSuccess &&
			n.ExecutedAt != nil &&
			n.ExecutedAt.After(now.Add(-time.Second)) &&
			n.UpdatedAt.After(now.Add(-time.Second))
	})).Return(nil)

	err := svc.MarkAsExecuted(id, true)
	assert.NoError(t, err)
}

func TestMarkAsExecuted_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := NewNotificationService(repo)

	id := uuid.New()
	repo.On("Find", id).Return(nil, model.ErrNotificationNotFound)

	err := svc.MarkAsExecuted(id, true)
	assert.ErrorIs(t, err, model.ErrNotificationNotFound)
}
