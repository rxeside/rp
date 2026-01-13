package activity

import (
	"context"

	"github.com/google/uuid"

	appdata "user/pkg/user/app/data"
	"user/pkg/user/app/service"
)

func NewUserServiceActivities(userService service.UserService) *UserServiceActivities {
	return &UserServiceActivities{userService: userService}
}

type UserServiceActivities struct {
	userService service.UserService
}

func (a *UserServiceActivities) FindUser(ctx context.Context, userID uuid.UUID) (appdata.User, error) {
	return a.userService.FindUser(ctx, userID)
}
