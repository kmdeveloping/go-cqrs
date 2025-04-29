package handler

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type DoThatCommandHandler struct {
	*cqrs.HandlerBase
}

var _ command.ICommandHandler[contracts.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d DoThatCommandHandler) Handle(command contracts.DoSomethingCommand) error {
	fmt.Println(command.Something)

	return cqrs.PublishEvent(d.Manager, contracts.SomeEvent{
		Name: command.Something,
	})
}
