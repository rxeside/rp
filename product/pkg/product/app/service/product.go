package service

import (
	"context"
	"time"

	"product/pkg/product/domain/model"

	"github.com/google/uuid"
)

type ProductService struct {
	repo model.ProductRepository
}

func NewProductService(repo model.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(_ context.Context, name string, price float64, quantity int) (uuid.UUID, error) {
	id, _ := s.repo.NextID()
	p := &model.Product{
		ID:        id,
		Name:      name,
		Price:     price,
		Quantity:  quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return id, s.repo.Store(p)
}

func (s *ProductService) Reserve(_ context.Context, id uuid.UUID, qty int) error {
	return s.repo.ReserveStock(id, qty)
}

func (s *ProductService) Release(_ context.Context, id uuid.UUID, qty int) error {
	return s.repo.ReleaseStock(id, qty)
}
