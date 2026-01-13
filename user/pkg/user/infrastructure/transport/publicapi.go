package transport

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"user/api/server/userpublicapi"
	appdata "user/pkg/user/app/data"
	"user/pkg/user/app/query"
	"user/pkg/user/app/service"
)

func NewUserInternalAPI(
	userQueryService query.UserQueryService,
	userService service.UserService,
) userpublicapi.UserPublicAPIServer {
	return &userInternalAPI{
		userQueryService: userQueryService,
		userService:      userService,
	}
}

type userInternalAPI struct {
	userQueryService query.UserQueryService
	userService      service.UserService

	userpublicapi.UnimplementedUserPublicAPIServer
}

func (u userInternalAPI) CreateUser(ctx context.Context, request *userpublicapi.CreateUserRequest) (*userpublicapi.CreateUserResponse, error) {
	userID, err := u.userService.CreateUser(ctx, appdata.User{
		Login:    request.Login,
		Email:    request.Email,
		Telegram: request.Telegram,
		Status:   0, // по умолчанию Blocked
	})
	if err != nil {
		return nil, err
	}

	return &userpublicapi.CreateUserResponse{
		UserID: userID.String(),
	}, nil
}

func (u userInternalAPI) UpdateUser(ctx context.Context, request *userpublicapi.UpdateUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.UserID)
	}

	update := appdata.UserUpdate{}
	if request.Status != userpublicapi.UserStatus(0) { // если не Blocked
		statusVal := int(request.Status)
		update.Status = &statusVal
	}
	if request.Email != nil {
		update.Email = request.Email
	}
	if request.Telegram != nil {
		update.Telegram = request.Telegram
	}

	err = u.userService.UpdateUser(ctx, userID, update)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (u userInternalAPI) FindUser(ctx context.Context, request *userpublicapi.FindUserRequest) (*userpublicapi.FindUserResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.UserID)
	}
	user, err := u.userQueryService.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user %q not found", request.UserID)
	}
	return &userpublicapi.FindUserResponse{
		UserID:   userID.String(),
		Status:   userpublicapi.UserStatus(user.Status), // #nosec: G115
		Login:    user.Login,
		Email:    user.Email,
		Telegram: user.Telegram,
	}, nil
}

func (u userInternalAPI) BlockUser(ctx context.Context, request *userpublicapi.BlockUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.UserID)
	}
	err = u.userService.BlockUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (u userInternalAPI) DeleteUser(ctx context.Context, request *userpublicapi.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.UserID)
	}
	err = u.userService.DeleteUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
