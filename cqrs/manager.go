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

// OPTIMIZATION: Dependency injection for clean API
var (
	currentManager     *Manager
	currentManagerMu   sync.RWMutex
	defaultManagerOnce sync.Once
)

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

// DEPENDENCY INJECTION: Set the current manager for all operations
// This should be called once during application startup
func SetManager(manager *Manager) {
	if manager == nil {
		panic("manager cannot be nil")
	}
	currentManagerMu.Lock()
	defer currentManagerMu.Unlock()
	currentManager = manager
}

// DEPENDENCY INJECTION: Get the current manager
// Creates a default manager if none is set (lazy initialization)
func GetManager() *Manager {
	currentManagerMu.RLock()
	if currentManager != nil {
		defer currentManagerMu.RUnlock()
		return currentManager
	}
	currentManagerMu.RUnlock()

	// Use double-checked locking for thread-safe lazy initialization
	currentManagerMu.Lock()
	defer currentManagerMu.Unlock()

	if currentManager == nil {
		defaultManagerOnce.Do(func() {
			currentManager = NewCqrsManager()
		})
	}

	return currentManager
}

// DEPENDENCY INJECTION: Check if a manager is explicitly set
func HasManagerSet() bool {
	currentManagerMu.RLock()
	defer currentManagerMu.RUnlock()
	return currentManager != nil
}

// DEPENDENCY INJECTION: Reset manager (useful for testing)
func ResetManager() {
	currentManagerMu.Lock()
	defer currentManagerMu.Unlock()
	currentManager = nil
	// Reset the once to allow creating a new default manager
	defaultManagerOnce = sync.Once{}
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

// CLEAN API: Clean registration method using current manager
func RegisterValidator[T command.ICommand](validator validator.IValidatorHandler[T]) error {
	return registerValidator(GetManager(), validator)
}

// CLEAN API: Clean registration method using current manager
func RegisterCommandHandler[T command.ICommand](handler command.ICommandHandler[T]) error {
	return registerCommandHandler(GetManager(), handler)
}

// CLEAN API: Clean registration method using current manager
func RegisterQueryHandler[T query.IQuery, R any](handler query.IQueryHandler[T, R]) error {
	return registerQueryHandler(GetManager(), handler)
}

// CLEAN API: Clean registration method using current manager
func RegisterEventHandler[T event.IEvent](handler event.IEventHandler[T]) error {
	return registerEventHandler(GetManager(), handler)
}

// Internal registration methods that work with specific manager instances
func registerValidator[T command.ICommand](m *Manager, validator validator.IValidatorHandler[T]) error {
	var zero T
	// Use pointer type for registration since validators now expect pointers
	typ := m.typeCache.getType(&zero)

	m.validatorsMu.Lock()
	defer m.validatorsMu.Unlock()
	m.validators[typ] = append(m.validators[typ], validator)

	return nil
}

func registerCommandHandler[T command.ICommand](m *Manager, handler command.ICommandHandler[T]) error {
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

func registerQueryHandler[T query.IQuery, R any](m *Manager, handler query.IQueryHandler[T, R]) error {
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

func registerEventHandler[T event.IEvent](m *Manager, handler event.IEventHandler[T]) error {
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

// CLEAN API: Add clean methods for adding decorators to current manager
func AddLoggingDecorator() *Manager {
	return GetManager().AddLoggingDecorator()
}

func AddMetricsDecorator() *Manager {
	return GetManager().AddMetricsDecorator()
}

func AddDecorator(decorator decorators.HandlerDecorator) *Manager {
	return GetManager().AddDecorator(decorator)
}

// CLEAN API: Add clean methods for metrics
func GetCommandCount() int64 {
	return GetManager().GetCommandCount()
}

func GetQueryCount() int64 {
	return GetManager().GetQueryCount()
}

func GetEventCount() int64 {
	return GetManager().GetEventCount()
}

func GetHandlerCounts() (commands, queries, events, validators int) {
	return GetManager().GetHandlerCounts()
}
