package cqrs

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/decorators"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

var mgr *Manager

type Manager struct {
	commandHandlers map[reflect.Type]any
	queryHandlers   map[reflect.Type]any
	eventHandlers   map[reflect.Type][]any
	validators      map[reflect.Type][]any
	decorators      []decorators.HandlerDecorator
	mu              sync.RWMutex
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

func (m *Manager) AddLoggingDecorator() *Manager {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loggingDecorator := decorators.LoggingDecorator(logger)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.decorators = append(m.decorators, loggingDecorator)

	return m
}

func (m *Manager) AddMetricsDecorator() *Manager {
	metricDecorator := decorators.MetricsDecorator()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.decorators = append(m.decorators, metricDecorator)

	return m
}

func (m *Manager) AddDecorator(decorator decorators.HandlerDecorator) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decorators = append(m.decorators, decorator)

	return m
}

func RegisterValidator[T command.ICommand](validator validator.IValidatorHandler[T]) {
	var zero T
	// Use pointer type for registration since validators now expect pointers
	typ := reflect.TypeOf(&zero)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.validators[typ] = append(mgr.validators[typ], validator)
}

func RegisterCommandHandler[T command.ICommand](handler command.ICommandHandler[T]) {
	var zero T
	// Use pointer type for registration since handlers now expect pointers
	typ := reflect.TypeOf(&zero)

	base := decorators.WrapCommandHandler(handler)
	decorated := decorators.WithDecorators(base, mgr.decorators...)

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
	decorated := decorators.WithDecorators(base, mgr.decorators...)

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
	decorated := decorators.WithDecorators(base, mgr.decorators...)
	unwrapped, ok := decorators.UnwrapAsEventHandler[T](decorated)
	if !ok {
		panic(fmt.Sprintf("failed to unwrap decorated handler for %T", typ))
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.eventHandlers[typ] = append(mgr.eventHandlers[typ], unwrapped)
}
