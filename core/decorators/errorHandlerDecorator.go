package decorators

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
)

type ErrorHandlerDecorator[T command.ICommand] struct {
	next command.ICommandHandler[T]
}

func (e *ErrorHandlerDecorator[T]) Handle(cmd T) error {
	fmt.Printf("ErrorHandlerDecorator[%T]\n", cmd)

	err := e.next.Handle(cmd)
	if err != nil {
		// do some handling stuff here
		return err
	}

	return nil
}

func UseErrorHandlerDecorator[T command.ICommand](handler command.ICommandHandler[T]) command.ICommandHandler[T] {
	return &ErrorHandlerDecorator[T]{next: handler}
}
