package integrationevent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"

	"user/pkg/user/domain/model"
	"user/pkg/user/infrastructure/temporal"
)

var errUnhandledDelivery = errors.New("unhandled delivery")

func NewAMQPTransport(logger logging.Logger, workflowService temporal.WorkflowService) AMQPTransport {
	return &amqpTransport{
		logger:          logger,
		workflowService: workflowService,
	}
}

type AMQPTransport interface {
	Handler() amqp.Handler
}

type amqpTransport struct {
	logger          logging.Logger
	workflowService temporal.WorkflowService
}

func (t *amqpTransport) Handler() amqp.Handler {
	return t.withLog(func(_ context.Context, delivery amqp.Delivery) error {
		switch delivery.Type {
		case model.UserCreated{}.Type():
			var e model.UserCreated
			err := json.Unmarshal(delivery.Body, &e)
			if err != nil {
				return err
			}
			t.logger.Info("received user_created event", "payload", e)
			return nil

		case model.UserUpdated{}.Type():
			var e model.UserUpdated
			err := json.Unmarshal(delivery.Body, &e)
			if err != nil {
				return err
			}
			t.logger.Info("received user_updated event", "payload", e)
			return nil // ← ТОЛЬКО ЛОГИРОВАНИЕ!

		case model.UserDeleted{}.Type():
			var e model.UserDeleted
			err := json.Unmarshal(delivery.Body, &e)
			if err != nil {
				return err
			}
			t.logger.Info("received user_deleted event", "payload", e)
			return nil

		default:
			return errUnhandledDelivery
		}
	})
}

func (t *amqpTransport) withLog(handler amqp.Handler) amqp.Handler {
	return func(ctx context.Context, delivery amqp.Delivery) error {
		l := t.logger.WithFields(logging.Fields{
			"routing_key":    delivery.RoutingKey,
			"correlation_id": delivery.CorrelationID,
			"content_type":   delivery.ContentType,
		})
		if delivery.ContentType != ContentType {
			l.Warning(errors.New("invalid content type"), "skipping")
			return nil
		}
		l = l.WithField("body", json.RawMessage(delivery.Body))

		start := time.Now()
		err := handler(ctx, delivery)
		l.WithField("duration", time.Since(start))

		if err != nil {
			if errors.Is(err, errUnhandledDelivery) {
				l.Info("unhandled delivery, skipping")
				return nil
			}
			l.Error(err, "failed to handle message")
		} else {
			l.Info("successfully handled message")
		}
		return err
	}
}
