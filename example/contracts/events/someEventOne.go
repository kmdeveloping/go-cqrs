package events

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/event"
)

type SomeEventOne struct {
	*event.EventBase
	Name string
}

var _ event.IEvent = (*SomeEventOne)(nil)

func (e *SomeEventOne) DoSomethingWithEventOne(event *SomeEventOne) error {
	fmt.Println("Executing event one: " + e.Name + " at " + e.ExecutionTime.String())
	return nil
}

func (e *SomeEventOne) RunSomeOtherEventForOne(event *SomeEventOne) error {
	fmt.Println("Running task for event one: " + e.Name)
	return nil
}
