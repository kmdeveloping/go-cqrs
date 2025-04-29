package cqrs

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/decorators"
	"reflect"
	"sync"
)

type CqrsManager struct {
	config   *CqrsConfiguration
	handlers map[reflect.Type]any
	mu       sync.RWMutex
}

func NewCqrsManager(config *CqrsConfiguration) *CqrsManager {
	return &CqrsManager{
		config:   config,
		handlers: make(map[reflect.Type]any),
	}
}

func RegisterHandler[T command.ICommand](mgr *CqrsManager, handler command.ICommandHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.handlers[typ] = handler
}

func Execute[T command.ICommand](mgr *CqrsManager, cmd T) error {
	typ := reflect.TypeOf(cmd)
	mgr.mu.RLock()
	handler, ok := mgr.handlers[typ]
	mgr.mu.RUnlock()
	if !ok {
		return fmt.Errorf("handler not found for type %v", typ)
	}

	typedHandler, ok := handler.(command.ICommandHandler[T])
	if !ok {
		return fmt.Errorf("handler type mismatch for %v", typ)
	}

	/// order decorators last is first to be called
	/// call stack: metrics >> errorHandler >> logger >> handler method
	if mgr.config.enableLoggingDecorator {
		typedHandler = decorators.UseLoggingDecorator(typedHandler)
	}

	if mgr.config.enableErrorHandlerDecorator {
		typedHandler = decorators.UseErrorHandlerDecorator(typedHandler)
	}

	if mgr.config.enableMetricsDecorator {
		typedHandler = decorators.UseExecutionTimeDecorator(typedHandler)
	}

	return typedHandler.Handle(cmd)
}
