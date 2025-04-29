package handler

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type DoThatCommandHandler struct{}

var _ command.ICommandHandler[contracts.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d *DoThatCommandHandler) Handle(command contracts.DoSomethingCommand) error {
	fmt.Println(command.Something)
	return nil
}
