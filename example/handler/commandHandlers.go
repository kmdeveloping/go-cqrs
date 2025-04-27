package handler

import (
	"errors"
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type DoThatCommandHandler struct {
	*handlers.BaseHandler
}

type DoThisCommandWithResultHandler struct {
	*DoThatCommandHandler
}

var _ handlers.ICommandHandler = (*DoThatCommandHandler)(nil)

func (d *DoThatCommandHandler) Execute(command command.ICommand) error {
	cmd, ok := command.(contracts.DoSomethingCommand)
	if !ok {
		return errors.New("invalid command type")
	}

	fmt.Println(cmd.Something)

	return nil
}

func (d *DoThisCommandWithResultHandler) ExecuteWithResult(cmd command.ICommandWithResult) error {
	c, ok := cmd.(contracts.DoSomethingWithResultCommand)
	if !ok {
		return errors.New("invalid command type")
	}

	c.Result = c.Something

	return nil
}
