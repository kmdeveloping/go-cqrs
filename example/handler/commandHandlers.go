package handler

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type DoThatCommandHandler struct {
}

var _ handlers.ICommandHandler[contracts.DoSomethingCommand] = (*DoThatCommandHandler)(nil)

func (d *DoThatCommandHandler) Execute(command contracts.DoSomethingCommand) error {
	fmt.Println(command.Something)
	return nil
}
