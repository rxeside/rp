package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"payment/pkg/payment/domain/model"
)

func NewPaymentRepository(ctx context.Context, client mysql.ClientContext) model.PaymentRepository {
	return &paymentRepository{
		ctx:    ctx,
		client: client,
	}
}

type paymentRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (p *paymentRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (p *paymentRepository) Store(payment *model.Payment) error {
	_, err := p.client.ExecContext(p.ctx,
		`
	INSERT INTO payment (payment_id, wallet_id, order_id, amount, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		wallet_id=VALUES(wallet_id),
		order_id=VALUES(order_id),
		amount=VALUES(amount),
		status=VALUES(status),
		updated_at=VALUES(updated_at),
		deleted_at=VALUES(deleted_at)
	`,
		payment.ID,
		payment.WalletID,
		payment.OrderID,
		payment.Amount,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
		toSQLNull(payment.DeletedAt),
	)
	return errors.WithStack(err)
}

func (p *paymentRepository) Find(id uuid.UUID) (*model.Payment, error) {
	paymentRow := struct {
		ID        uuid.UUID           `db:"payment_id"`
		WalletID  uuid.UUID           `db:"wallet_id"`
		OrderID   uuid.UUID           `db:"order_id"`
		Amount    float64             `db:"amount"`
		Status    int                 `db:"status"`
		CreatedAt time.Time           `db:"created_at"`
		UpdatedAt time.Time           `db:"updated_at"`
		DeletedAt sql.Null[time.Time] `db:"deleted_at"`
	}{}

	err := p.client.GetContext(
		p.ctx,
		&paymentRow,
		`SELECT payment_id, wallet_id, order_id, amount, status, created_at, updated_at, deleted_at FROM payment WHERE payment_id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrPaymentNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Payment{
		ID:        paymentRow.ID,
		WalletID:  paymentRow.WalletID,
		OrderID:   paymentRow.OrderID,
		Amount:    paymentRow.Amount,
		Status:    model.PaymentStatus(paymentRow.Status),
		CreatedAt: paymentRow.CreatedAt,
		UpdatedAt: paymentRow.UpdatedAt,
		DeletedAt: fromSQLNull(paymentRow.DeletedAt),
	}, nil
}

func (p *paymentRepository) Remove(id uuid.UUID) error {
	now := time.Now()
	_, err := p.client.ExecContext(p.ctx,
		`UPDATE payment SET deleted_at = ? WHERE payment_id = ?`,
		now,
		id,
	)
	return errors.WithStack(err)
}
