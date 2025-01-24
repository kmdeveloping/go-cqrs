package commands

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/command"
)

type DoSomethingCommand struct {
	*command.CommandBase
	CustomerNumber string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)

// non-public handler execute function
func (c *DoSomethingCommand) Handler(command *DoSomethingCommand) error {
	fmt.Println(command.CustomerNumber)
	return nil
}
