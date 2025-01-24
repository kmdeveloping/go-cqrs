package cqrs

import (
	"reflect"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/decorators"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type ICqrsManager interface {
	Execute(T command.ICommand) error
	Get(T query.IQuery) error
	Publish(T event.IEvent) error
	Validate(T validator.IValidator) error
}

type CqrsManager struct {
	config *CqrsConfiguration
}

var _ ICqrsManager = (*CqrsManager)(nil)

func NewCqrsManager(config *CqrsConfiguration) ICqrsManager {
	return &CqrsManager{config: config}
}

func (m *CqrsManager) Execute(T command.ICommand) error {
	handler, err := m.Setup(T)
	if err != nil {
		return err
	}

	if err = handler.Run(T); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) Get(T query.IQuery) error {
	handler, err := m.Setup(T)
	if err != nil {
		return err
	}

	if err = handler.Get(T); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) Publish(T event.IEvent) error {
	handler, err := m.Setup(T)
	if err != nil {
		return err
	}

	if err = handler.Publish(T); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) Validate(T validator.IValidator) error {
	handler, err := m.Setup(T)
	if err != nil {
		return err
	}

	if err = handler.Validate(T); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) Setup(T any) (handlers.IHandler, error) {
	handler, err := m.config.Registry.Resolve(reflect.TypeOf(T))
	if err != nil {
		return nil, err
	}

	if m.config.EnableLoggingDecorator {
		handler = decorators.UseLoggingDecorator(handler)
	}

	return handler, nil
}
