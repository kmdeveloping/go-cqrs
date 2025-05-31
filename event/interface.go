package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type IEvent any

type IEventHandler[T IEvent] interface {
	Handle(ctx context.Context, event T) error
}

type Base struct {
	ExecutionTime  time.Time
	CorrelationUid uuid.UUID
	MetaData       string
}

var _ IEvent = (*Base)(nil)
