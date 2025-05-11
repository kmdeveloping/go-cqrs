package handlers

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type DoThatCommandHandler struct{}

var _ command.ICommandHandler[commands.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d DoThatCommandHandler) Handle(command commands.DoSomethingCommand) error {
	fmt.Println(command.Something)

	if command.Something != "" {
		command.Result = "It's done!"
	} else {
		command.Result = "Nothing to do!"
	}
	return cqrs.PublishEvent(events.SomeEvent{
		Name: command.Something,
	})
}
