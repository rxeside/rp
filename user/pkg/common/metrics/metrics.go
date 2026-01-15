package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// EventsProcessed считает количество обработанных событий из RabbitMQ
var EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "app_events_processed_total",
	Help: "The total number of processed events",
}, []string{"event_type", "status"}) // status: success, error, unhandled
