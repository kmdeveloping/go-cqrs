package handlers

import (
	"context"
	"log"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type SomeEventHandler struct{}

func (s SomeEventHandler) Handle(ctx context.Context, event events.SomeEvent) error {
	log.Println("event handler trigger for " + event.Name)
	return nil
}

var _ event.IEventHandler[events.SomeEvent] = (*SomeEventHandler)(nil)
