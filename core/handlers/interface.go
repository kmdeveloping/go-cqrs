package handlers

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type IHandler interface {
	ICommandHandler
	IQueryHandler[any]
	IEventHandler
	IValidatorHandler
}

type ICommandHandler interface {
	Run(TCommand command.ICommand) error
}

type IQueryHandler[TResult any] interface {
	Get(TQuery query.IQuery) (TResult, error)
}

type IEventHandler interface {
	Publish(TEvent event.IEvent) error
}

type IValidatorHandler interface {
	Validate(TValidator validator.IValidator) error
}

type BaseHandler struct{}

var _ IHandler = (*BaseHandler)(nil)

func (h *BaseHandler) Run(TCommand command.ICommand) error {
	return nil
}

func (h *BaseHandler) Get(TQuery query.IQuery) (any, error) {
	return nil, nil
}

func (h *BaseHandler) Publish(TEvent event.IEvent) error {
	return nil
}

func (h *BaseHandler) Validate(TValidator validator.IValidator) error {
	return nil
}
