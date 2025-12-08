package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "user/pkg/user/application/model"
)

type UserQueryService interface {
	FindUser(ctx context.Context, userID uuid.UUID) (*appmodel.User, error)
}
