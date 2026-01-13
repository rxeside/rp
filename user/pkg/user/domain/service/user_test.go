package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"user/pkg/common/domain"
	"user/pkg/user/domain/model"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) Store(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Find(spec model.FindSpec) (*model.User, error) {
	args := m.Called(spec)
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) HardDelete(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestCreateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "testuser"
	userID := uuid.New()

	repo.On("Find", mock.Anything).Return(nil, model.ErrUserNotFound)
	repo.On("NextID").Return(userID, nil)
	repo.On("Store", mock.Anything).Return(nil)
	dispatcher.On("Dispatch", mock.AnythingOfType("*model.UserCreated")).Return(nil)

	svc := NewUserService(repo, dispatcher)
	id, err := svc.CreateUser(login)

	assert.NoError(t, err)
	assert.Equal(t, userID, id)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestCreateUser_LoginExists(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	login := "existinguser"
	existingUser := &model.User{Login: login}

	repo.On("Find", mock.Anything).Return(existingUser, nil)

	svc := NewUserService(repo, dispatcher)
	_, err := svc.CreateUser(login)

	assert.ErrorIs(t, err, model.ErrUserLoginAlreadyUsed)
	repo.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	userID := uuid.New()
	existingUser := &model.User{UserID: userID, Login: "test"}

	repo.On("Find", mock.Anything).Return(existingUser, nil)
	repo.On("Store", mock.Anything).Return(nil)
	dispatcher.On("Dispatch", mock.AnythingOfType("*model.UserUpdated")).Return(nil)

	svc := NewUserService(repo, dispatcher)
	err := svc.UpdateUser(userID, struct {
		Status   *model.UserStatus
		Email    *string
		Telegram *string
	}{
		Status: func() *model.UserStatus { s := model.Active; return &s }(),
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)

	userID := uuid.New()
	existingUser := &model.User{UserID: userID}

	repo.On("Find", mock.Anything).Return(existingUser, nil)
	repo.On("Store", mock.Anything).Return(nil)
	dispatcher.On("Dispatch", mock.AnythingOfType("*model.UserDeleted")).Return(nil)

	svc := NewUserService(repo, dispatcher)
	err := svc.DeleteUser(userID, false)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}
