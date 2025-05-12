package handlers

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type DoThatCommandHandler struct{}

// Make sure handler implements the interface with pointer commands
var _ command.ICommandHandler[commands.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d DoThatCommandHandler) Handle(command *commands.DoSomethingCommand) error {
	log.Println(command.Something)

	command.Result = "Hello from DoThatCommandHandler"

	return cqrs.PublishEvent(events.SomeEvent{
		Name: command.Something,
	})
}
