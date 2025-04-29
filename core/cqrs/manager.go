package cqrs

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
	"reflect"
	"sync"
)

type Manager struct {
	config          *Configuration
	commandHandlers map[reflect.Type]any
	queryHandlers   map[reflect.Type]any
	eventHandlers   map[reflect.Type][]any
	validators      map[reflect.Type][]any
	mu              sync.RWMutex
}

func NewCqrsManager(config *Configuration) *Manager {
	return &Manager{
		config:          config,
		commandHandlers: make(map[reflect.Type]any),
		queryHandlers:   make(map[reflect.Type]any),
		eventHandlers:   make(map[reflect.Type][]any),
		validators:      make(map[reflect.Type][]any),
	}
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
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.commandHandlers[typ] = handler
}

func RegisterQueryHandler[T query.IQuery, R any](bus *Manager, handler query.IQueryHandler[T, R]) {
	var zero T
	typ := reflect.TypeOf(zero)
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.queryHandlers[typ] = handler
}

func RegisterEventHandler[T event.IEvent](bus *Manager, handler event.IEventHandler[T]) {
	var zero T
	typ := reflect.TypeOf(zero)

	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.eventHandlers[typ] = append(bus.eventHandlers[typ], handler)
}
