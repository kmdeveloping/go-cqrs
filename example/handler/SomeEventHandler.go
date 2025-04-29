package handler

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type SomeEventHandler struct{}

func (s SomeEventHandler) Handle(event contracts.SomeEvent) error {
	fmt.Println("event handler trigger for " + event.Name)
	return nil
}

var _ event.IEventHandler[contracts.SomeEvent] = (*SomeEventHandler)(nil)
