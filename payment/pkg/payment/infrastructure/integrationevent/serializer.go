package integrationevent

import (
	"encoding/json"
	"payment/pkg/payment/domain/model"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	if e, ok := event.(*model.UserCreated); ok {
		b, _ := json.Marshal(e)
		return string(b), nil
	}
	return "", nil
}
