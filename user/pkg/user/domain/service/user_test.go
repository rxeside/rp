package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	commonevent "user/pkg/common/event"
	"user/pkg/user/domain/model"
)

const (
	testUserLogin = "testuser"
	testUserEmail = "test@example.com"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) Store(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Find(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Remove(id uuid.UUID) error {
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

func newUser(id uuid.UUID, login, email string) *model.User {
	now := time.Now()
	return &model.User{
		ID:        id,
		Login:     login,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCreateUser_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDispatcher := new(MockEventDispatcher)

	login := testUserLogin
	email := testUserEmail
	userID := uuid.New()

	userRepo.On("NextID").Return(userID, nil)
	userRepo.On("Store", mock.MatchedBy(func(user *model.User) bool {
		return user.ID == userID &&
			user.Login == login &&
			user.Email == email &&
			!user.CreatedAt.IsZero() &&
			user.UpdatedAt.Equal(user.CreatedAt) &&
			user.DeletedAt == nil
	})).Return(nil)

	eventDispatcher.On("Dispatch", mock.MatchedBy(func(e model.UserCreated) bool {
		return e.UserID == userID && e.Login == login && e.Email == email
	})).Return(nil)

	svc := NewUserService(userRepo, eventDispatcher)

	id, err := svc.CreateUser(login, email)

	assert.NoError(t, err)
	assert.Equal(t, userID, id)
	userRepo.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}

func TestCreateUser_RepoError(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	login := testUserLogin
	email := testUserEmail
	userID := uuid.New()

	userRepo.On("NextID").Return(userID, nil)
	userRepo.On("Store", mock.Anything).Return(errors.New("db down"))

	svc := NewUserService(userRepo, eventDisp)

	_, err := svc.CreateUser(login, email)
	assert.Error(t, err)
	userRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestCreateUser_EventDispatchError(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	login := testUserLogin
	email := testUserEmail
	userID := uuid.New()

	userRepo.On("NextID").Return(userID, nil)
	userRepo.On("Store", mock.Anything).Return(nil)
	eventDisp.On("Dispatch", mock.Anything).Return(errors.New("kafka unreachable"))

	svc := NewUserService(userRepo, eventDisp)

	_, err := svc.CreateUser(login, email)
	assert.Error(t, err)
	userRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	login := testUserLogin
	email := testUserEmail

	user := newUser(userID, "olduser", "old@example.com")
	userRepo.On("Find", userID).Return(user, nil)
	userRepo.On("Store", mock.MatchedBy(func(u *model.User) bool {
		return u.Login == login && u.Email == email && u.UpdatedAt.After(u.CreatedAt)
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.UserUpdated) bool {
		return e.UserID == userID
	})).Return(nil)

	svc := NewUserService(userRepo, eventDisp)

	err := svc.UpdateUser(userID, login, email)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestUpdateUser_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	login := testUserLogin
	email := testUserEmail

	userRepo.On("Find", userID).Return(nil, model.ErrUserNotFound)

	svc := NewUserService(userRepo, eventDisp)

	err := svc.UpdateUser(userID, login, email)
	assert.ErrorIs(t, err, model.ErrUserNotFound)
	userRepo.AssertExpectations(t)
}

func TestRemoveUser_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()

	user := newUser(userID, testUserLogin, testUserEmail)
	userRepo.On("Find", userID).Return(user, nil)
	userRepo.On("Store", mock.MatchedBy(func(u *model.User) bool {
		return u.ID == userID && u.DeletedAt != nil
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.UserRemoved) bool {
		return e.UserID == userID
	})).Return(nil)

	svc := NewUserService(userRepo, eventDisp)

	err := svc.RemoveUser(userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveUser_NotFound_Idempotent(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	userRepo.On("Find", userID).Return(nil, model.ErrUserNotFound)

	svc := NewUserService(userRepo, eventDisp)

	err := svc.RemoveUser(userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestRemoveUser_FindError(t *testing.T) {
	userRepo := new(MockUserRepository)
	eventDisp := new(MockEventDispatcher)

	userID := uuid.New()
	userRepo.On("Find", userID).Return(nil, errors.New("db timeout"))

	svc := NewUserService(userRepo, eventDisp)

	err := svc.RemoveUser(userID)
	assert.Error(t, err)
}
