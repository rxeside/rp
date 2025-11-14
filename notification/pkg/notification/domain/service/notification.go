package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "notification/pkg/common/event"
	"notification/pkg/notification/domain/model"
)

var (
	ErrNotificationNotFound      = errors.New("notification not found")
	ErrInvalidNotificationStatus = errors.New("invalid notification status")
)

type NotificationService interface {
	CreateNotification(userID, orderID uuid.UUID, message string) (uuid.UUID, error)
	RemoveNotification(notificationID uuid.UUID) error
	SetStatus(notificationID uuid.UUID, status string) error
}

func NewNotificationService(repo model.NotificationRepository, dispatcher commonevent.Dispatcher) NotificationService {
	return &notificationService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type notificationService struct {
	repo       model.NotificationRepository
	dispatcher commonevent.Dispatcher
}

func (n *notificationService) CreateNotification(userID, orderID uuid.UUID, message string) (uuid.UUID, error) {
	notificationID, err := n.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = n.repo.Store(&model.Notification{
		ID:        notificationID,
		UserID:    userID,
		OrderID:   orderID,
		Status:    "pending",
		Message:   message,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return notificationID, n.dispatcher.Dispatch(model.NotificationCreated{
		NotificationID: notificationID,
		UserID:         userID,
		OrderID:        orderID,
		Status:         "pending",
		Message:        message,
	})
}

func (n *notificationService) RemoveNotification(notificationID uuid.UUID) error {
	notification, err := n.repo.Find(notificationID)
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	notification.DeletedAt = &now
	notification.UpdatedAt = now

	if err = n.repo.Store(notification); err != nil {
		return err
	}

	return n.dispatcher.Dispatch(model.NotificationRemoved{
		NotificationID: notificationID,
	})
}

func (n *notificationService) SetStatus(notificationID uuid.UUID, status string) error {
	notification, err := n.repo.Find(notificationID)
	if err != nil {
		return err
	}

	oldStatus := notification.Status

	if !n.isValidStatusTransition(oldStatus, status) {
		return ErrInvalidNotificationStatus
	}

	notification.Status = status
	notification.UpdatedAt = time.Now()

	if err = n.repo.Store(notification); err != nil {
		return err
	}

	return n.dispatcher.Dispatch(model.NotificationStatusChanged{
		NotificationID: notificationID,
		From:           oldStatus,
		To:             status,
	})
}

func (n *notificationService) isValidStatusTransition(from, to string) bool {
	switch from {
	case "pending":
		return to == "sent" || to == "failed" || to == "cancelled"
	case "sent", "failed", "cancelled":
		return false
	default:
		return false
	}
}
