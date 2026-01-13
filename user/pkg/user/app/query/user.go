package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "user/pkg/user/app/data"
)

type UserQueryService interface {
	FindUser(ctx context.Context, userID uuid.UUID) (*appmodel.User, error)
}
