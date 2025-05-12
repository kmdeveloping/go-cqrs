package handlers

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type SomeOtherEventHandler struct{}

func (s SomeOtherEventHandler) Handle(event events.SomeEvent) error {
	log.Println("SomeOtherEventHandler")
	return nil
}

var _ event.IEventHandler[events.SomeEvent] = (*SomeOtherEventHandler)(nil)
