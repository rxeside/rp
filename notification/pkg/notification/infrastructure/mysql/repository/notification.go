package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"notification/pkg/notification/domain/model"
)

func NewNotificationRepository(ctx context.Context, client mysql.ClientContext) model.NotificationRepository {
	return &notificationRepository{
		ctx:    ctx,
		client: client,
	}
}

type notificationRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *notificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *notificationRepository) Store(notification model.Notification) error {
	payloadBytes, err := json.Marshal(notification.Payload)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = r.client.ExecContext(r.ctx,
		`INSERT INTO notification (id, payload, executed_at, created_at, updated_at, status)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		 payload = VALUES(payload),
		 executed_at = VALUES(executed_at),
		 updated_at = VALUES(updated_at),
		 status = VALUES(status)`,
		notification.ID,
		payloadBytes,
		toSQLNullTime(notification.ExecutedAt),
		notification.CreatedAt,
		notification.UpdatedAt,
		toSQLNullString(notification.Status),
	)
	return errors.WithStack(err)
}

func (r *notificationRepository) Find(id uuid.UUID) (*model.Notification, error) {
	row := struct {
		ID         uuid.UUID       `db:"id"`
		Payload    json.RawMessage `db:"payload"`
		ExecutedAt sql.NullTime    `db:"executed_at"`
		CreatedAt  time.Time       `db:"created_at"`
		UpdatedAt  time.Time       `db:"updated_at"`
		Status     sql.NullString  `db:"status"`
	}{}

	err := r.client.GetContext(r.ctx, &row,
		`SELECT id, payload, executed_at, created_at, updated_at, status FROM notification WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrNotificationNotFound)
		}
		return nil, errors.WithStack(err)
	}

	var payload model.NotificationPayload
	if err := json.Unmarshal(row.Payload, &payload); err != nil {
		return nil, errors.WithStack(err)
	}

	return &model.Notification{
		ID:         row.ID,
		Payload:    payload,
		ExecutedAt: fromSQLNullTime(row.ExecutedAt),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		Status:     fromSQLNullString(row.Status),
	}, nil
}

func toSQLNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func fromSQLNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func toSQLNullString(s *model.NotificationStatus) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(*s), Valid: true}
}

func fromSQLNullString(ns sql.NullString) *model.NotificationStatus {
	if !ns.Valid {
		return nil
	}
	status := model.NotificationStatus(ns.String)
	return &status
}
