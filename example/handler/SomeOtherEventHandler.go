package handler

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type SomeOtherEventHandler struct{}

func (s SomeOtherEventHandler) Handle(event contracts.SomeEvent) error {
	fmt.Println("SomeOtherEventHandler")
	return nil
}

var _ event.IEventHandler[contracts.SomeEvent] = (*SomeOtherEventHandler)(nil)
