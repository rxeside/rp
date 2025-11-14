package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "product/pkg/common/event"
	"product/pkg/product/domain/model"
)

type Product interface {
	CreateProduct(name string, price float64) (uuid.UUID, error)
	UpdateProduct(productID uuid.UUID, name string, price float64) error
	RemoveProduct(productID uuid.UUID) error
}

func NewProductService(repo model.ProductRepository, dispatcher commonevent.Dispatcher) Product {
	return &productService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type productService struct {
	repo       model.ProductRepository
	dispatcher commonevent.Dispatcher
}

func (p productService) CreateProduct(name string, price float64) (uuid.UUID, error) {
	productID, err := p.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = p.repo.Store(&model.Product{
		ID:        productID,
		Name:      name,
		Price:     price,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return productID, p.dispatcher.Dispatch(model.ProductCreated{
		ProductID: productID,
		Name:      name,
		Price:     price,
	})
}

func (p productService) UpdateProduct(productID uuid.UUID, name string, price float64) error {
	product, err := p.repo.Find(productID)
	if err != nil {
		return err
	}

	product.Name = name
	product.Price = price
	product.UpdatedAt = time.Now()

	if err = p.repo.Store(product); err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.ProductUpdated{
		ProductID: productID,
	})
}

func (p productService) RemoveProduct(productID uuid.UUID) error {
	product, err := p.repo.Find(productID)
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	product.DeletedAt = &now
	product.UpdatedAt = now

	if err = p.repo.Store(product); err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.ProductRemoved{
		ProductID: productID,
	})
}
