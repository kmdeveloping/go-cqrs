package event

import (
	"time"

	"github.com/google/uuid"
)

type IEvent interface{}

type IEventHandler[T IEvent] interface {
	Publish(T) error
}

type EventBase struct {
	ExecutionTime  time.Time
	CorrelationUid uuid.UUID
	MetaData       string
}

var _ IEvent = (*EventBase)(nil)
