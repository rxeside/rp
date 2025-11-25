package service

import (
	"github.com/stretchr/testify/mock"

	commonevent "payment/pkg/common/event"
)

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event commonevent.Event) error {
	args := m.Called(event)
	return args.Error(0)
}
