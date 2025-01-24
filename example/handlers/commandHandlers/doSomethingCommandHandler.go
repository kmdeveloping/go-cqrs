package commandHandlers

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/example/contracts/commands"
)

// DoSomethingCommandHandler extends base command handler
type DoSomethingCommandHandler struct {
	command.BaseCommandHandler[commands.DoSomethingCommand]
}

// non-public handler execute function
func (c *DoSomethingCommandHandler) execute(command *commands.DoSomethingCommand) error {
	fmt.Println(command.CustomerNumber)
	return nil
}

// NewDoSomethingCommand getter to return handler for command
func NewDoSomethingCommand() *DoSomethingCommandHandler {
	handler := &DoSomethingCommandHandler{}
	handler.Execute = handler.execute
	return handler
}
