package events

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/event"
)

type SomeEventTwo struct {
	*event.EventBase
	Name string
}

var _ event.IEvent = (*SomeEventTwo)(nil)

func (e *SomeEventTwo) RunEventTwo(event *SomeEventTwo) error {
	fmt.Println("Executed event two: " + e.Name)
	return nil
}
