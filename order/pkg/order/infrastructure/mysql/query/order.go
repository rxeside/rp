package query

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"order/pkg/order/app/data"
	"order/pkg/order/app/query"
	"order/pkg/order/domain/model"
)

func NewOrderQueryService(client mysql.ClientContext) query.OrderQueryService {
	return &orderQueryService{
		client: client,
	}
}

type orderQueryService struct {
	client mysql.ClientContext
}

func (o *orderQueryService) FindUser(ctx context.Context, orderID uuid.UUID) (*data.Order, error) {
	orderRow := struct {
		ID         uuid.UUID           `db:"order_id"`
		CustomerID uuid.UUID           `db:"customer_id"`
		Status     int                 `db:"status"`
		CreatedAt  time.Time           `db:"created_at"`
		UpdatedAt  time.Time           `db:"updated_at"`
		DeletedAt  sql.Null[time.Time] `db:"deleted_at"`
	}{}

	err := o.client.GetContext(
		ctx,
		&orderRow,
		`SELECT order_id, customer_id, status, created_at, updated_at, deleted_at FROM order WHERE order_id = ?`,
		orderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrOrderNotFound)
		}
		return nil, errors.WithStack(err)
	}

	items, err := o.loadOrderItems(ctx, orderRow.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &data.Order{
		ID:         orderRow.ID,
		CustomerID: orderRow.CustomerID,
		Status:     data.OrderStatus(orderRow.Status),
		Items:      items,
		CreatedAt:  orderRow.CreatedAt,
		UpdatedAt:  orderRow.UpdatedAt,
		DeletedAt:  fromSQLNull(orderRow.DeletedAt),
	}, nil
}

func (o *orderQueryService) loadOrderItems(ctx context.Context, orderID uuid.UUID) ([]data.OrderItem, error) {
	var itemRows []struct {
		OrderID    uuid.UUID `db:"order_id"`
		ProductID  uuid.UUID `db:"product_id"`
		Count      int       `db:"count"`
		TotalPrice float64   `db:"total_price"`
	}

	err := o.client.SelectContext(
		ctx,
		&itemRows,
		`SELECT order_id, product_id, count, total_price FROM order_item WHERE order_id = ?`,
		orderID,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := make([]data.OrderItem, len(itemRows))
	for i, row := range itemRows {
		items[i] = data.OrderItem{
			OrderID:    row.OrderID,
			ProductID:  row.ProductID,
			Count:      row.Count,
			TotalPrice: row.TotalPrice,
		}
	}

	return items, nil
}

func fromSQLNull[T any](v sql.Null[T]) *T {
	if v.Valid {
		return &v.V
	}
	return nil
}
