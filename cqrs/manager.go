package cqrs

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/decorators"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

// OPTIMIZATION: Removed global singleton pattern to eliminate race conditions and improve testability
// var mgr *Manager // REMOVED

// typeCache stores reflection information to avoid expensive reflect.TypeOf calls in hot paths
type typeCache struct {
	mu    sync.RWMutex
	cache map[interface{}]reflect.Type
}

func newTypeCache() *typeCache {
	return &typeCache{
		cache: make(map[interface{}]reflect.Type),
	}
}

// OPTIMIZATION: Cache reflection data to improve performance in hot paths
func (tc *typeCache) getType(v interface{}) reflect.Type {
	// Fast path: try to get from cache with read lock
	tc.mu.RLock()
	if typ, exists := tc.cache[v]; exists {
		tc.mu.RUnlock()
		return typ
	}
	tc.mu.RUnlock()

	// Slow path: compute and cache the type
	typ := reflect.TypeOf(v)
	tc.mu.Lock()
	tc.cache[v] = typ
	tc.mu.Unlock()

	return typ
}

type Manager struct {
	commandHandlers map[reflect.Type]any
	queryHandlers   map[reflect.Type]any
	eventHandlers   map[reflect.Type][]any
	validators      map[reflect.Type][]any
	decorators      []decorators.HandlerDecorator

	// OPTIMIZATION: Use RWMutex for better read performance and separate locks for different operations
	handlersMu   sync.RWMutex
	validatorsMu sync.RWMutex
	decoratorsMu sync.RWMutex

	// OPTIMIZATION: Add type cache to avoid expensive reflection calls
	typeCache *typeCache

	// OPTIMIZATION: Add atomic counter for metrics
	commandCount int64
	queryCount   int64
	eventCount   int64
}

// OPTIMIZATION: Factory function instead of global singleton
func NewCqrsManager() *Manager {
	return &Manager{
		commandHandlers: make(map[reflect.Type]any),
		queryHandlers:   make(map[reflect.Type]any),
		eventHandlers:   make(map[reflect.Type][]any),
		validators:      make(map[reflect.Type][]any),
		typeCache:       newTypeCache(),
	}
}

// OPTIMIZATION: Improved thread safety with specific mutex
func (m *Manager) AddLoggingDecorator() *Manager {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	loggingDecorator := decorators.LoggingDecorator(logger)

	m.decoratorsMu.Lock()
	defer m.decoratorsMu.Unlock()
	m.decorators = append(m.decorators, loggingDecorator)

	return m
}

// OPTIMIZATION: Improved thread safety with specific mutex
func (m *Manager) AddMetricsDecorator() *Manager {
	metricDecorator := decorators.MetricsDecorator()

	m.decoratorsMu.Lock()
	defer m.decoratorsMu.Unlock()
	m.decorators = append(m.decorators, metricDecorator)

	return m
}

// OPTIMIZATION: Improved thread safety with specific mutex
func (m *Manager) AddDecorator(decorator decorators.HandlerDecorator) *Manager {
	m.decoratorsMu.Lock()
	defer m.decoratorsMu.Unlock()
	m.decorators = append(m.decorators, decorator)

	return m
}

// OPTIMIZATION: Convert to function that accepts Manager instance instead of global function
func RegisterValidator[T command.ICommand](m *Manager, validator validator.IValidatorHandler[T]) error {
	var zero T
	// Use pointer type for registration since validators now expect pointers
	typ := m.typeCache.getType(&zero)

	m.validatorsMu.Lock()
	defer m.validatorsMu.Unlock()
	m.validators[typ] = append(m.validators[typ], validator)

	return nil
}

// OPTIMIZATION: Convert to function that accepts Manager instance with proper error handling
func RegisterCommandHandler[T command.ICommand](m *Manager, handler command.ICommandHandler[T]) error {
	var zero T
	// Use pointer type for registration since handlers now expect pointers
	typ := m.typeCache.getType(&zero)

	// OPTIMIZATION: Get decorators with read lock to avoid blocking other operations
	m.decoratorsMu.RLock()
	decoratorsCopy := make([]decorators.HandlerDecorator, len(m.decorators))
	copy(decoratorsCopy, m.decorators)
	m.decoratorsMu.RUnlock()

	base := decorators.WrapCommandHandler(handler)
	decorated := decorators.WithDecorators(base, decoratorsCopy...)

	unwrapped, ok := decorators.UnwrapAsCommandHandler[T](decorated)
	if !ok {
		// OPTIMIZATION: Return error instead of panic for better error handling
		return fmt.Errorf("failed to unwrap decorated handler for %T", typ)
	}

	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.commandHandlers[typ] = unwrapped

	return nil
}

// OPTIMIZATION: Convert to function that accepts Manager instance with proper error handling
func RegisterQueryHandler[T query.IQuery, R any](m *Manager, handler query.IQueryHandler[T, R]) error {
	var zero T
	typ := m.typeCache.getType(zero)

	// OPTIMIZATION: Get decorators with read lock
	m.decoratorsMu.RLock()
	decoratorsCopy := make([]decorators.HandlerDecorator, len(m.decorators))
	copy(decoratorsCopy, m.decorators)
	m.decoratorsMu.RUnlock()

	base := decorators.WrapQueryHandler(handler)
	decorated := decorators.WithDecorators(base, decoratorsCopy...)

	unwrapped, ok := decorators.UnwrapAsQueryHandler[T, R](decorated)
	if !ok {
		// OPTIMIZATION: Return error instead of panic for better error handling
		return fmt.Errorf("failed to unwrap decorated handler for %T", typ)
	}

	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.queryHandlers[typ] = unwrapped

	return nil
}

// OPTIMIZATION: Convert to function that accepts Manager instance with proper error handling
func RegisterEventHandler[T event.IEvent](m *Manager, handler event.IEventHandler[T]) error {
	var zero T
	typ := m.typeCache.getType(zero)

	// OPTIMIZATION: Get decorators with read lock
	m.decoratorsMu.RLock()
	decoratorsCopy := make([]decorators.HandlerDecorator, len(m.decorators))
	copy(decoratorsCopy, m.decorators)
	m.decoratorsMu.RUnlock()

	base := decorators.WrapEventHandler(handler)
	decorated := decorators.WithDecorators(base, decoratorsCopy...)
	unwrapped, ok := decorators.UnwrapAsEventHandler[T](decorated)
	if !ok {
		// OPTIMIZATION: Return error instead of panic for better error handling
		return fmt.Errorf("failed to unwrap decorated handler for %T", typ)
	}

	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.eventHandlers[typ] = append(m.eventHandlers[typ], unwrapped)

	return nil
}

// OPTIMIZATION: Add metrics methods for monitoring
func (m *Manager) GetCommandCount() int64 {
	return atomic.LoadInt64(&m.commandCount)
}

func (m *Manager) GetQueryCount() int64 {
	return atomic.LoadInt64(&m.queryCount)
}

func (m *Manager) GetEventCount() int64 {
	return atomic.LoadInt64(&m.eventCount)
}

// OPTIMIZATION: Add method to get handler counts for monitoring
func (m *Manager) GetHandlerCounts() (commands, queries, events, validators int) {
	m.handlersMu.RLock()
	commands = len(m.commandHandlers)
	queries = len(m.queryHandlers)
	events = len(m.eventHandlers)
	m.handlersMu.RUnlock()

	m.validatorsMu.RLock()
	validators = len(m.validators)
	m.validatorsMu.RUnlock()

	return
}
