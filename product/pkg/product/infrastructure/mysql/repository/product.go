package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"product/pkg/product/domain/model"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) model.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *productRepository) Store(p *model.Product) error {
	// Обновлен запрос: добавлено поле deleted_at
	_, err := r.db.Exec(`
		INSERT INTO products (id, name, price, quantity, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name),
			price=VALUES(price),
			quantity=VALUES(quantity),
			updated_at=VALUES(updated_at),
			deleted_at=VALUES(deleted_at)
	`, p.ID.String(), p.Name, p.Price, p.Quantity, p.CreatedAt, p.UpdatedAt, toSQLNullTime(p.DeletedAt))
	return errors.WithStack(err)
}

func (r *productRepository) Find(id uuid.UUID) (*model.Product, error) {
	// Используем вспомогательную структуру для сканирования Nullable полей
	var row struct {
		ID        string       `db:"id"`
		Name      string       `db:"name"`
		Price     float64      `db:"price"`
		Quantity  int          `db:"quantity"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt time.Time    `db:"updated_at"`
		DeletedAt sql.NullTime `db:"deleted_at"`
	}

	// Обновлен запрос: добавлено поле deleted_at
	err := r.db.QueryRowx(`
		SELECT id, name, price, quantity, created_at, updated_at, deleted_at 
		FROM products WHERE id = ?`, id.String()).StructScan(&row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrProductNotFound
		}
		return nil, errors.WithStack(err)
	}

	return &model.Product{
		ID:        uuid.MustParse(row.ID),
		Name:      row.Name,
		Price:     row.Price,
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: fromSQLNullTime(row.DeletedAt),
	}, nil
}

// --- ДОБАВЛЕН МЕТОД REMOVE ---
func (r *productRepository) Remove(id uuid.UUID) error {
	// Реализуем Soft Delete
	_, err := r.db.Exec(`UPDATE products SET deleted_at = ? WHERE id = ?`, time.Now(), id.String())
	return errors.WithStack(err)
}

// -----------------------------

func (r *productRepository) ReserveStock(id uuid.UUID, quantity int) error {
	res, err := r.db.Exec(`UPDATE products SET quantity = quantity - ? WHERE id = ? AND quantity >= ?`, quantity, id.String(), quantity)
	if err != nil {
		return errors.WithStack(err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return model.ErrInsufficientStock
	}
	return nil
}

func (r *productRepository) ReleaseStock(id uuid.UUID, quantity int) error {
	_, err := r.db.Exec(`UPDATE products SET quantity = quantity + ? WHERE id = ?`, quantity, id.String())
	return errors.WithStack(err)
}

// Вспомогательные функции для конвертации *time.Time <-> sql.NullTime

func toSQLNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}

func fromSQLNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
