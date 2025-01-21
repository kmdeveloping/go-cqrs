package commands

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/command"
)

type GetSomethingCommand struct {
	CustomerNumber string
}

func (c *GetSomethingCommand) Execute(command *GetSomethingCommand) error {
	fmt.Println(command.CustomerNumber)
	return nil
}

func NewGetSomethingCommandHandler() command.CommandHandler[*GetSomethingCommand] {
	return &GetSomethingCommand{}
}
