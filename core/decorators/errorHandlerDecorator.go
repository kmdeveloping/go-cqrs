package decorators

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type ErrorHandlerDecorator struct {
	next handlers.IHandler
}

var _ handlers.IHandler = (*ErrorHandlerDecorator)(nil)

func (e ErrorHandlerDecorator) Execute(TCommand command.ICommand) error {
	err := e.next.Execute(TCommand)
	if err != nil {
		// do some handling stuff here
		return err
	}

	return nil
}

func (e ErrorHandlerDecorator) ExecuteWithResult(TCommandWithResult command.ICommandWithResult) error {
	err := e.next.Execute(TCommandWithResult)
	if err != nil {
		// do some handling stuff here
		return err
	}

	return nil
}

func (e ErrorHandlerDecorator) Get(TQuery query.IQuery) (any, error) {
	result, err := e.next.Get(TQuery)
	if err != nil {
		// do some handling stuff here
		return nil, err
	}

	return result, nil
}

func (e ErrorHandlerDecorator) Publish(TEvent event.IEvent) error {
	err := e.next.Publish(TEvent)
	if err != nil {
		// do some handling stuff here
		return err
	}

	return nil
}

func (e ErrorHandlerDecorator) Validate(TValidator validator.IValidator) error {
	err := e.next.Validate(TValidator)
	if err != nil {
		// do some handling stuff here
		return err
	}

	return nil
}

func UseErrorHandlerDecorator(handler handlers.IHandler) handlers.IHandler {
	return &ErrorHandlerDecorator{handler}
}
