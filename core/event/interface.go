package event

import (
	"time"

	"github.com/google/uuid"
)

type IEvent interface{}

type EventBase struct {
	ExecutionTime  time.Time
	CorrelationUid uuid.UUID
	MetaData       string
}

var _ IEvent = (*EventBase)(nil)
