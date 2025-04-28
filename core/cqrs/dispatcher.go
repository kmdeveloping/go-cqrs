package cqrs

import (
	"errors"
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/decorators"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/registry"
	"github.com/kmdeveloping/go-cqrs/core/validator"
	"reflect"
)

type ICqrsManager interface {
	Execute(cmd T) error
	Get(T query.IQuery) (any, error)
	Publish(T event.IEvent) error
	Validate(T validator.IValidator) error
	UseLoggingDecorator() ICqrsManager
	UseMetricsDecorator() ICqrsManager
	UseErrorHandlerDecorator() ICqrsManager
}

type CqrsManager struct {
	config *CqrsConfiguration
}

var _ ICqrsManager = (*CqrsManager)(nil)

func NewCqrsManager(config *CqrsConfiguration) ICqrsManager {
	return &CqrsManager{config: config}
}

func (m *CqrsManager) UseLoggingDecorator() ICqrsManager {
	m.config.enableLoggingDecorator = true
	return m
}

func (m *CqrsManager) UseMetricsDecorator() ICqrsManager {
	m.config.enableMetricsDecorator = true
	return m
}

func (m *CqrsManager) UseErrorHandlerDecorator() ICqrsManager {
	m.config.enableErrorHandlerDecorator = true
	return m
}

func (m *CqrsManager) Execute(contract any) error {

	t := reflect.TypeOf(contract).Out(0)
	handler, ok := m.config.Handlers.Handlers.Load(t).(handlers.ICommandHandler[t])
	if !ok {
		return errors.New(fmt.Sprintf("handler not found for type %v", reflect.TypeOf(T)))
	}

	if m.config.enableLoggingDecorator {
		handler = decorators.UseLoggingDecorator(handler)
	}

	if m.config.enableMetricsDecorator {
		handler = decorators.UseExecutionTimeDecorator(handler)
	}

	if m.config.enableErrorHandlerDecorator {
		handler = decorators.UseErrorHandlerDecorator(handler)
	}

	if err = handler.Execute(contract); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) Get(T query.IQuery) (any, error) {
	handler, err := m.setup(T)
	if err != nil {
		return nil, err
	}

	response, err := handler.Get(T)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (m *CqrsManager) Publish(T event.IEvent) error {
	handler, err := m.setup(T)
	if err != nil {
		return err
	}

	if err = handler.Publish(T); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) Validate(T validator.IValidator) error {
	handler, err := m.setup(T)
	if err != nil {
		return err
	}

	if err = handler.Validate(T); err != nil {
		return err
	}

	return nil
}

func (m *CqrsManager) setup(T any) (handlers.IHandler, error) {
	handler, err := registry.LoadCommand[T](&m.config.Handlers)
	if err != nil {
		return nil, err
	}

	if m.config.enableLoggingDecorator {
		handler = decorators.UseLoggingDecorator(handler)
	}

	if m.config.enableMetricsDecorator {
		handler = decorators.UseExecutionTimeDecorator(handler)
	}

	if m.config.enableErrorHandlerDecorator {
		handler = decorators.UseErrorHandlerDecorator(handler)
	}

	return handler, nil
}
