package temporal

import (
	"context"
	"product/pkg/product/app/service"

	"github.com/google/uuid"
)

type ProductActivities struct {
	svc *service.ProductService
}

func NewProductActivities(svc *service.ProductService) *ProductActivities {
	return &ProductActivities{svc: svc}
}

func (a *ProductActivities) ReserveProduct(ctx context.Context, productID string, quantity int) error {
	id, err := uuid.Parse(productID)
	if err != nil {
		return err
	}
	return a.svc.Reserve(ctx, id, quantity)
}

func (a *ProductActivities) ReleaseProduct(ctx context.Context, productID string, quantity int) error {
	id, err := uuid.Parse(productID)
	if err != nil {
		return err
	}
	return a.svc.Release(ctx, id, quantity)
}
