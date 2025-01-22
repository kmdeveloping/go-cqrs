package commands

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/command"
)

type GetSomethingCommand struct {
	CustomerNumber string
}

type GetSomethingCommandHandler struct {
	command.BaseCommandHandler[GetSomethingCommand]
}

func (c *GetSomethingCommandHandler) execute(command *GetSomethingCommand) error {
	fmt.Println(command.CustomerNumber)
	return nil
}

func NewGetSomethingCommand() *GetSomethingCommandHandler {
	handler := &GetSomethingCommandHandler{}
	handler.Execute = handler.execute
	return handler
}
