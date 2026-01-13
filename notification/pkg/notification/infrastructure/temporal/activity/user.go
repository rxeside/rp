package activity

import (
	"context"

	"github.com/google/uuid"

	appdata "notification/pkg/notification/app/data"
	"notification/pkg/notification/app/service"
)

func NewUserActivities(userService service.UserService) *UserActivities {
	return &UserActivities{userService: userService}
}

type UserActivities struct {
	userService service.UserService
}

func (a *UserActivities) FindUser(ctx context.Context, userID uuid.UUID) (appdata.User, error) {
	return a.userService.FindUser(ctx, userID)
}

func (a *UserActivities) SetUserStatus(ctx context.Context, userID uuid.UUID, status int) error {
	return a.userService.SetUserStatus(ctx, userID, status)
}

func (a *UserActivities) StoreUser(ctx context.Context, user appdata.User) (uuid.UUID, error) {
	return a.userService.StoreUser(ctx, user)
}
