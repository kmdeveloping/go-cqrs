package commandHandlers

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	. "github.com/kmdeveloping/go-cqrs/example/contracts/commands"
)

type DoSomethingCommandHandler struct {
	command.BaseCommandHandler[DoSomethingCommand] `di.inject:"cmdHandler"`
}

func (c *DoSomethingCommandHandler) execute(command *DoSomethingCommand) error {
	fmt.Println(command.CustomerNumber)
	return nil
}

func NewDoSomethingCommand() *DoSomethingCommandHandler {
	handler := &DoSomethingCommandHandler{}
	handler.Execute = handler.execute
	return handler
}
