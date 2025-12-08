package integrationevent

import (
	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(_ outbox.Event) (string, error) {
	return "", nil
}
