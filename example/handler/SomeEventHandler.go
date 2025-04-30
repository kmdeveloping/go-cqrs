package handler

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type SomeEventHandler struct{}

func (s SomeEventHandler) Handle(event events.SomeEvent) error {
	fmt.Println("event handler trigger for " + event.Name)
	return nil
}

var _ event.IEventHandler[events.SomeEvent] = (*SomeEventHandler)(nil)
