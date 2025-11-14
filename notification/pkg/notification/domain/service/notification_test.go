package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notification/pkg/common/event"
	"notification/pkg/notification/domain/model"
)

const (
	testMessage   = "test message"
	statusSent    = "sent"
	statusPending = "pending"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockNotificationRepository) Store(notification *model.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) Find(id uuid.UUID) (*model.Notification, error) {
	args := m.Called(id)
	if notification, ok := args.Get(0).(*model.Notification); ok {
		return notification, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNotificationRepository) Remove(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event event.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func newPendingNotification(id, userID, orderID uuid.UUID) *model.Notification {
	now := time.Now()
	return &model.Notification{
		ID:        id,
		UserID:    userID,
		OrderID:   orderID,
		Status:    statusPending,
		Message:   testMessage,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newSentNotification(id, userID, orderID uuid.UUID) *model.Notification {
	notification := newPendingNotification(id, userID, orderID)
	notification.Status = statusSent
	return notification
}

func TestCreateNotification_Success(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDispatcher := new(MockEventDispatcher)

	userID := uuid.New()
	orderID := uuid.New()
	notificationID := uuid.New()

	notificationRepo.On("NextID").Return(notificationID, nil)
	notificationRepo.On("Store", mock.MatchedBy(func(notification *model.Notification) bool {
		return notification.ID == notificationID &&
			notification.UserID == userID &&
			notification.OrderID == orderID &&
			notification.Status == statusPending &&
			notification.Message == testMessage &&
			!notification.CreatedAt.IsZero() &&
			notification.UpdatedAt.Equal(notification.CreatedAt)
	})).Return(nil)

	eventDispatcher.On("Dispatch", mock.MatchedBy(func(e model.NotificationCreated) bool {
		return e.NotificationID == notificationID && e.UserID == userID && e.OrderID == orderID && e.Status == statusPending && e.Message == testMessage
	})).Return(nil)

	svc := NewNotificationService(notificationRepo, eventDispatcher)

	id, err := svc.CreateNotification(userID, orderID, testMessage)

	assert.NoError(t, err)
	assert.Equal(t, notificationID, id)
	notificationRepo.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}

func TestCreateNotification_RepoError(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	orderID := uuid.New()
	notificationID := uuid.New()

	notificationRepo.On("NextID").Return(notificationID, nil)
	notificationRepo.On("Store", mock.Anything).Return(errors.New("db down"))

	svc := NewNotificationService(notificationRepo, eventDisp)

	_, err := svc.CreateNotification(userID, orderID, testMessage)
	assert.Error(t, err)
	notificationRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestCreateNotification_EventDispatchError(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	orderID := uuid.New()
	notificationID := uuid.New()

	notificationRepo.On("NextID").Return(notificationID, nil)
	notificationRepo.On("Store", mock.Anything).Return(nil)
	eventDisp.On("Dispatch", mock.Anything).Return(errors.New("kafka unreachable"))

	svc := NewNotificationService(notificationRepo, eventDisp)

	_, err := svc.CreateNotification(userID, orderID, testMessage)
	assert.Error(t, err)
	notificationRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveNotification_Success(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	notificationID := uuid.New()
	userID := uuid.New()
	orderID := uuid.New()
	notification := newPendingNotification(notificationID, userID, orderID)

	notificationRepo.On("Find", notificationID).Return(notification, nil)
	notificationRepo.On("Store", mock.MatchedBy(func(n *model.Notification) bool {
		return n.ID == notificationID && n.DeletedAt != nil
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.NotificationRemoved) bool {
		return e.NotificationID == notificationID
	})).Return(nil)

	svc := NewNotificationService(notificationRepo, eventDisp)

	err := svc.RemoveNotification(notificationID)
	assert.NoError(t, err)
	notificationRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveNotification_NotFound_Idempotent(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	notificationID := uuid.New()
	notificationRepo.On("Find", notificationID).Return(nil, ErrNotificationNotFound)

	svc := NewNotificationService(notificationRepo, eventDisp)

	err := svc.RemoveNotification(notificationID)
	assert.NoError(t, err)
	notificationRepo.AssertExpectations(t)
}

func TestRemoveNotification_FindError(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	notificationID := uuid.New()
	notificationRepo.On("Find", notificationID).Return(nil, errors.New("db timeout"))

	svc := NewNotificationService(notificationRepo, eventDisp)

	err := svc.RemoveNotification(notificationID)
	assert.Error(t, err)
}

func TestSetStatus_ValidTransition_PendingToSent(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	notificationID := uuid.New()
	userID := uuid.New()
	orderID := uuid.New()
	notification := newPendingNotification(notificationID, userID, orderID)

	notificationRepo.On("Find", notificationID).Return(notification, nil)
	notificationRepo.On("Store", mock.MatchedBy(func(n *model.Notification) bool {
		return n.Status == statusSent
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.NotificationStatusChanged) bool {
		return e.NotificationID == notificationID && e.From == statusPending && e.To == statusSent
	})).Return(nil)

	svc := NewNotificationService(notificationRepo, eventDisp)

	err := svc.SetStatus(notificationID, statusSent)
	assert.NoError(t, err)
}

func TestSetStatus_InvalidTransition_SentToPending(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	notificationID := uuid.New()
	userID := uuid.New()
	orderID := uuid.New()
	notification := newSentNotification(notificationID, userID, orderID)

	notificationRepo.On("Find", notificationID).Return(notification, nil)

	svc := NewNotificationService(notificationRepo, eventDisp)

	err := svc.SetStatus(notificationID, statusPending)
	assert.ErrorIs(t, err, ErrInvalidNotificationStatus)
	notificationRepo.AssertExpectations(t)
}

func TestSetStatus_NotificationNotFound(t *testing.T) {
	notificationRepo := new(MockNotificationRepository)
	eventDisp := new(MockEventDispatcher)

	notificationID := uuid.New()
	notificationRepo.On("Find", notificationID).Return(nil, ErrNotificationNotFound)

	svc := NewNotificationService(notificationRepo, eventDisp)

	err := svc.SetStatus(notificationID, "cancelled")
	assert.ErrorIs(t, err, ErrNotificationNotFound)
}

func TestIsValidStatusTransition_Matrix(t *testing.T) {
	svc := &notificationService{}

	tests := []struct {
		from  string
		to    string
		valid bool
		desc  string
	}{
		{statusPending, statusSent, true, "pending → sent"},
		{statusPending, "failed", true, "pending → failed"},
		{statusPending, "cancelled", true, "pending → cancelled"},
		{statusPending, statusPending, false, "pending → pending (no-op)"},
		{statusPending, "delivered", false, "pending → delivered (invalid)"},

		{statusSent, statusPending, false, "sent → pending"},
		{statusSent, "failed", false, "sent → failed"},
		{statusSent, "delivered", false, "sent → delivered"},
		{statusSent, statusSent, false, "sent → sent"},

		{"failed", statusPending, false, "failed → pending"},
		{"failed", statusSent, false, "failed → sent"},
		{"cancelled", statusPending, false, "cancelled → pending"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.valid, svc.isValidStatusTransition(tt.from, tt.to))
		})
	}
}
