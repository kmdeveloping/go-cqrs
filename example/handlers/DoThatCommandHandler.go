package handlers

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type DoThatCommandHandler struct{}

// We register the handler for pointer type commands
// But we don't need to assert the interface implementation since we made ICommandHandler accept any types
var _ command.ICommandHandler[commands.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d DoThatCommandHandler) Handle(command commands.DoSomethingCommand) error {
	fmt.Println(command.Something)

	return cqrs.PublishEvent(events.SomeEvent{
		Name: command.Something,
	})
}
