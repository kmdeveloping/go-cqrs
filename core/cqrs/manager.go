package cqrs

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/decorators"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

var mgr *Manager

type Manager struct {
	commandHandlers   map[reflect.Type]any
	queryHandlers     map[reflect.Type]any
	eventHandlers     map[reflect.Type][]any
	validators        map[reflect.Type][]any
	defaultDecorators []decorators.HandlerDecorator
	mu                sync.RWMutex
}

func NewCqrsManager() *Manager {
	mgr = &Manager{
		commandHandlers: make(map[reflect.Type]any),
		queryHandlers:   make(map[reflect.Type]any),
		eventHandlers:   make(map[reflect.Type][]any),
		validators:      make(map[reflect.Type][]any),
	}

	return mgr
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

func (m *Manager) AddLoggingDecorator() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loggingDecorator := decorators.LoggingDecorator(logger)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDecorators = append(m.defaultDecorators, loggingDecorator)
}

func (m *Manager) AddMetricsDecorator() {
	metricDecorator := decorators.MetricsDecorator()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDecorators = append(m.defaultDecorators, metricDecorator)
}

func (m *Manager) AddErrorHandlerDecorator() {
	errorHandlerDecorator := decorators.ErrorHandlerDecorator()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDecorators = append(m.defaultDecorators, errorHandlerDecorator)
}

func RegisterValidator[T command.ICommand](validator validator.IValidatorHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.validators[typ] = append(mgr.validators[typ], validator)
}

func RegisterCommandHandler[T command.ICommand](handler command.ICommandHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	base := decorators.WrapCommandHandler(handler)
	decorated := decorators.WithDecorators(base, mgr.defaultDecorators...)

	unwrapped, ok := decorators.UnwrapAsCommandHandler[T](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.commandHandlers[typ] = unwrapped
}

func RegisterQueryHandler[T query.IQuery, R any](handler query.IQueryHandler[T, R]) {
	var zero T
	typ := reflect.TypeOf(zero)

	base := decorators.WrapQueryHandler(handler)
	decorated := decorators.WithDecorators(base, mgr.defaultDecorators...)

	unwrapped, ok := decorators.UnwrapAsQueryHandler[T, R](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.queryHandlers[typ] = unwrapped
}

func RegisterEventHandler[T event.IEvent](handler event.IEventHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	base := decorators.WrapEventHandler(handler)
	decorated := decorators.WithDecorators(base, mgr.defaultDecorators...)

	unwrapped, ok := decorators.UnwrapAsEventHandler[T](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.eventHandlers[typ] = append(mgr.eventHandlers[typ], unwrapped)
}
