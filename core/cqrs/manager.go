package cqrs

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/decorators"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
	"log"
	"os"
	"reflect"
	"sync"
)

type Manager struct {
	commandHandlers   map[reflect.Type]any
	queryHandlers     map[reflect.Type]any
	eventHandlers     map[reflect.Type][]any
	validators        map[reflect.Type][]any
	defaultDecorators []decorators.HandlerDecorator
	mu                sync.RWMutex
}

func NewCqrsManager() *Manager {
	return &Manager{
		commandHandlers: make(map[reflect.Type]any),
		queryHandlers:   make(map[reflect.Type]any),
		eventHandlers:   make(map[reflect.Type][]any),
		validators:      make(map[reflect.Type][]any),
	}
}

func (m *Manager) UseDefaultDecorators() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loggingDecorator := decorators.LoggingDecorator(logger)
	metricDecorator := decorators.MetricsDecorator()
	errorHandlerDecorator := decorators.ErrorHandlerDecorator()

	/// indexing layers decorators as 0 => most outer decorator to N => most inner decorator before func call
	defaults := []decorators.HandlerDecorator{
		metricDecorator,
		loggingDecorator,
		errorHandlerDecorator,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDecorators = append(m.defaultDecorators, defaults...)
}

func RegisterValidator[T command.ICommand](bus *Manager, validator validator.IValidatorHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.validators[typ] = append(bus.validators[typ], validator)
}

func RegisterCommandHandler[T command.ICommand](bus *Manager, handler command.ICommandHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	base := decorators.WrapCommandHandler(handler)
	decorated := decorators.WithDecorators(base, bus.defaultDecorators...)

	unwrapped, ok := decorators.UnwrapAsCommandHandler[T](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.commandHandlers[typ] = unwrapped
}

func RegisterQueryHandler[T query.IQuery, R any](bus *Manager, handler query.IQueryHandler[T, R]) {
	var zero T
	typ := reflect.TypeOf(zero)

	base := decorators.WrapQueryHandler(handler)
	decorated := decorators.WithDecorators(base, bus.defaultDecorators...)

	unwrapped, ok := decorators.UnwrapAsQueryHandler[T, R](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.queryHandlers[typ] = unwrapped
}

func RegisterEventHandler[T event.IEvent](bus *Manager, handler event.IEventHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	base := decorators.WrapEventHandler(handler)
	decorated := decorators.WithDecorators(base, bus.defaultDecorators...)

	unwrapped, ok := decorators.UnwrapAsEventHandler[T](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.eventHandlers[typ] = append(bus.eventHandlers[typ], unwrapped)
}
