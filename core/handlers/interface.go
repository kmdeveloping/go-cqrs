package handlers

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type IHandler interface {
	ICommandHandler
	ICommandWithResultHandler
	IQueryHandler
	IEventHandler
	IValidatorHandler
}

type ICommandHandler interface {
	Execute(TCommand command.ICommand) error
}

type ICommandWithResultHandler interface {
	ExecuteWithResult(TCommandWithResult command.ICommandWithResult) error
}

type IQueryHandler interface {
	Get(TQuery query.IQuery) (interface{}, error)
}

type IEventHandler interface {
	Publish(TEvent event.IEvent) error
}

type IValidatorHandler interface {
	Validate(TValidator validator.IValidator) error
}
