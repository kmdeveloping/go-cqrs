package handlers

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type IHandler interface {
	ICommandHandler[command.ICommand]
	IQueryHandler[query.IQuery, any]
	IEventHandler[event.IEvent]
	IValidatorHandler[validator.IValidator]
}

type ICommandHandler[TCommand command.ICommand] interface {
	Execute(TCommand) error
}

type IQueryHandler[TQuery query.IQuery, TResult any] interface {
	Get(TQuery) (TResult, error)
}

type IEventHandler[TEvent event.IEvent] interface {
	Publish(TEvent) error
}

type IValidatorHandler[TValidator validator.IValidator] interface {
	Validate(TValidator) error
}
