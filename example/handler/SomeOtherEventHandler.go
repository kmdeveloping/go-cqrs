package handler

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type SomeOtherEventHandler struct{}

func (s SomeOtherEventHandler) Handle(event events.SomeEvent) error {
	fmt.Println("SomeOtherEventHandler")
	return nil
}

var _ event.IEventHandler[events.SomeEvent] = (*SomeOtherEventHandler)(nil)
