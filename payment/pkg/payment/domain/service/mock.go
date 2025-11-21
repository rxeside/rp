package service

import (
	"payment/pkg/common/event"

	"github.com/stretchr/testify/mock"
)

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event event.Event) error {
	args := m.Called(event)
	return args.Error(0)
}
