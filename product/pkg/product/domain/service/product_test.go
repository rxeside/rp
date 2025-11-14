package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	commonevent "product/pkg/common/event"
	"product/pkg/product/domain/model"
)

const testName = "TestProduct"

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockProductRepository) Store(product *model.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Find(id uuid.UUID) (*model.Product, error) {
	args := m.Called(id)
	if product, ok := args.Get(0).(*model.Product); ok {
		return product, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) Remove(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event commonevent.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func newProduct(id uuid.UUID, name string, price float64) *model.Product {
	now := time.Now()
	return &model.Product{
		ID:        id,
		Name:      name,
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCreateProduct_Success(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDispatcher := new(MockEventDispatcher)

	name := testName
	price := 99.99
	productID := uuid.New()

	productRepo.On("NextID").Return(productID, nil)
	productRepo.On("Store", mock.MatchedBy(func(product *model.Product) bool {
		return product.ID == productID &&
			product.Name == name &&
			product.Price == price &&
			!product.CreatedAt.IsZero() &&
			product.UpdatedAt.Equal(product.CreatedAt) &&
			product.DeletedAt == nil
	})).Return(nil)

	eventDispatcher.On("Dispatch", mock.MatchedBy(func(e model.ProductCreated) bool {
		return e.ProductID == productID && e.Name == name && e.Price == price
	})).Return(nil)

	svc := NewProductService(productRepo, eventDispatcher)

	id, err := svc.CreateProduct(name, price)

	assert.NoError(t, err)
	assert.Equal(t, productID, id)
	productRepo.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}

func TestCreateProduct_RepoError(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	name := testName
	price := 99.99
	productID := uuid.New()

	productRepo.On("NextID").Return(productID, nil)
	productRepo.On("Store", mock.Anything).Return(errors.New("db down"))

	svc := NewProductService(productRepo, eventDisp)

	_, err := svc.CreateProduct(name, price)
	assert.Error(t, err)
	productRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestCreateProduct_EventDispatchError(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	name := testName
	price := 99.99
	productID := uuid.New()

	productRepo.On("NextID").Return(productID, nil)
	productRepo.On("Store", mock.Anything).Return(nil)
	eventDisp.On("Dispatch", mock.Anything).Return(errors.New("kafka unreachable"))

	svc := NewProductService(productRepo, eventDisp)

	_, err := svc.CreateProduct(name, price)
	assert.Error(t, err)
	productRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestUpdateProduct_Success(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	productID := uuid.New()
	name := testName
	price := 149.99

	product := newProduct(productID, "Old Product", 49.99)
	productRepo.On("Find", productID).Return(product, nil)
	productRepo.On("Store", mock.MatchedBy(func(p *model.Product) bool {
		return p.Name == name && p.Price == price && p.UpdatedAt.After(p.CreatedAt)
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.ProductUpdated) bool {
		return e.ProductID == productID
	})).Return(nil)

	svc := NewProductService(productRepo, eventDisp)

	err := svc.UpdateProduct(productID, name, price)
	assert.NoError(t, err)
	productRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestUpdateProduct_ProductNotFound(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	productID := uuid.New()
	name := testName
	price := 149.99

	productRepo.On("Find", productID).Return(nil, model.ErrProductNotFound)

	svc := NewProductService(productRepo, eventDisp)

	err := svc.UpdateProduct(productID, name, price)
	assert.ErrorIs(t, err, model.ErrProductNotFound)
	productRepo.AssertExpectations(t)
}

func TestRemoveProduct_Success(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	productID := uuid.New()

	product := newProduct(productID, "To Remove", 29.99)
	productRepo.On("Find", productID).Return(product, nil)
	productRepo.On("Store", mock.MatchedBy(func(p *model.Product) bool {
		return p.ID == productID && p.DeletedAt != nil
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.ProductRemoved) bool {
		return e.ProductID == productID
	})).Return(nil)

	svc := NewProductService(productRepo, eventDisp)

	err := svc.RemoveProduct(productID)
	assert.NoError(t, err)
	productRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveProduct_NotFound_Idempotent(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	productID := uuid.New()
	productRepo.On("Find", productID).Return(nil, model.ErrProductNotFound)

	svc := NewProductService(productRepo, eventDisp)

	err := svc.RemoveProduct(productID)
	assert.NoError(t, err)
	productRepo.AssertExpectations(t)
}

func TestRemoveProduct_FindError(t *testing.T) {
	productRepo := new(MockProductRepository)
	eventDisp := new(MockEventDispatcher)

	productID := uuid.New()
	productRepo.On("Find", productID).Return(nil, errors.New("db timeout"))

	svc := NewProductService(productRepo, eventDisp)

	err := svc.RemoveProduct(productID)
	assert.Error(t, err)
}
