package handler

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

type DoThatCommandHandler struct {
	*cqrs.HandlerBase
}

var _ command.ICommandHandler[commands.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d DoThatCommandHandler) Handle(command commands.DoSomethingCommand) error {
	fmt.Println(command.Something)

	return cqrs.PublishEvent(d.Manager, events.SomeEvent{
		Name: command.Something,
	})
}
