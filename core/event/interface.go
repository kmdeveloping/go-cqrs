package event

import (
	"time"

	"github.com/google/uuid"
)

type IEvent interface{}

type IEventHandler[T IEvent] interface {
	Handle(event T) error
}

type Base struct {
	ExecutionTime  time.Time
	CorrelationUid uuid.UUID
	MetaData       string
}

var _ IEvent = (*Base)(nil)
