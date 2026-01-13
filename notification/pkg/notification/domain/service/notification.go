package service

import (
	"time"

	"github.com/google/uuid"

	"notification/pkg/notification/domain/model"
)

type NotificationService interface {
	CreateNotification(payload model.NotificationPayload) (uuid.UUID, error)
	MarkAsExecuted(id uuid.UUID, success bool) error
}

func NewNotificationService(repo model.NotificationRepository) NotificationService {
	return &notificationService{
		repo: repo,
	}
}

type notificationService struct {
	repo model.NotificationRepository
}

func (s *notificationService) CreateNotification(payload model.NotificationPayload) (uuid.UUID, error) {
	id, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	now := time.Now()
	notif := model.Notification{
		ID:        id,
		Payload:   payload,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Store(notif); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (s *notificationService) MarkAsExecuted(id uuid.UUID, success bool) error {
	notif, err := s.repo.Find(id)
	if err != nil {
		return err
	}

	executedAt := time.Now()
	var status model.NotificationStatus
	if success {
		status = model.StatusSuccess
	} else {
		status = model.StatusFailed
	}

	notif.ExecutedAt = &executedAt
	notif.Status = &status
	notif.UpdatedAt = executedAt

	return s.repo.Store(*notif)
}
