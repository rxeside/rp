package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "user/pkg/common/event"
	"user/pkg/user/domain/model"
)

type User interface {
	CreateUser(login, email string) (uuid.UUID, error)
	UpdateUser(userID uuid.UUID, login, email string) error
	RemoveUser(userID uuid.UUID) error
}

func NewUserService(repo model.UserRepository, dispatcher commonevent.Dispatcher) User {
	return &userService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type userService struct {
	repo       model.UserRepository
	dispatcher commonevent.Dispatcher
}

func (u userService) CreateUser(login, email string) (uuid.UUID, error) {
	userID, err := u.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = u.repo.Store(&model.User{
		ID:        userID,
		Login:     login,
		Email:     email,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return userID, u.dispatcher.Dispatch(model.UserCreated{
		UserID: userID,
		Login:  login,
		Email:  email,
	})
}

func (u userService) UpdateUser(userID uuid.UUID, login, email string) error {
	user, err := u.repo.Find(userID)
	if err != nil {
		return err
	}

	user.Login = login
	user.Email = email
	user.UpdatedAt = time.Now()

	if err = u.repo.Store(user); err != nil {
		return err
	}

	return u.dispatcher.Dispatch(model.UserUpdated{
		UserID: userID,
	})
}

func (u userService) RemoveUser(userID uuid.UUID) error {
	user, err := u.repo.Find(userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil
		}
		return err
	}

	now := time.Now()
	user.DeletedAt = &now
	user.UpdatedAt = now

	if err = u.repo.Store(user); err != nil {
		return err
	}

	return u.dispatcher.Dispatch(model.UserRemoved{
		UserID: userID,
	})
}
