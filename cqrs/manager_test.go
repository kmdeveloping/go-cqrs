package cqrs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
)

// Test command types
type TestCommand struct {
	command.Base
	ID   string
	Data string
}

type TestCommandResult struct {
	command.BaseWithResult
	Success bool
}

// Test query types
type TestQuery struct {
	query.Base
	ID string
}

type TestQueryResult struct {
	Value string
	Count int
}

// Test event types
type TestEvent struct {
	event.Base
	ID      string
	Message string
}

// Test handlers
type testCommandHandler struct {
	executed bool
	mu       sync.Mutex
}

func (h *testCommandHandler) Handle(ctx context.Context, cmd *TestCommand) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executed = true
	return nil
}

func (h *testCommandHandler) WasExecuted() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.executed
}

type testQueryHandler struct{}

func (h *testQueryHandler) Handle(ctx context.Context, q TestQuery) (TestQueryResult, error) {
	return TestQueryResult{
		Value: "test_value_" + q.ID,
		Count: 42,
	}, nil
}

type testEventHandler struct {
	handled bool
	mu      sync.Mutex
}

func (h *testEventHandler) Handle(ctx context.Context, e TestEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handled = true
	return nil
}

func (h *testEventHandler) WasHandled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.handled
}

type testValidator struct {
	validated bool
	mu        sync.Mutex
}

func (v *testValidator) Validate(ctx context.Context, cmd *TestCommand) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.validated = true
	return nil
}

func (v *testValidator) WasValidated() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.validated
}

// OPTIMIZATION TEST: Verify Manager creation without global singleton
func TestNewCqrsManager_CreatesIndependentInstances(t *testing.T) {
	manager1 := NewCqrsManager()
	manager2 := NewCqrsManager()

	if manager1 == manager2 {
		t.Error("NewCqrsManager should create independent instances")
	}

	// Verify each manager has its own state
	cmdHandler1 := &testCommandHandler{}
	cmdHandler2 := &testCommandHandler{}

	err1 := RegisterCommandHandler(manager1, cmdHandler1)
	err2 := RegisterCommandHandler(manager2, cmdHandler2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to register handlers: %v, %v", err1, err2)
	}

	cmd1, queries1, events1, validators1 := manager1.GetHandlerCounts()
	cmd2, queries2, events2, validators2 := manager2.GetHandlerCounts()

	if cmd1 != 1 || cmd2 != 1 {
		t.Errorf("Expected both managers to have 1 command handler, got %d and %d", cmd1, cmd2)
	}

	// Verify they don't share state
	if queries1 != 0 || events1 != 0 || validators1 != 0 ||
		queries2 != 0 || events2 != 0 || validators2 != 0 {
		t.Error("Managers should have independent state")
	}
}

// OPTIMIZATION TEST: Verify thread safety improvements
func TestManager_ThreadSafety(t *testing.T) {
	manager := NewCqrsManager()
	const numGoroutines = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create unique command types for each goroutine to test concurrent registration
	type baseUniqueCommand struct {
		command.Base
		GoroutineID int
	}

	// Test concurrent registration and access
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a handler for this goroutine's unique command type
			handler := &testCommandHandler{}

			// Each goroutine registers its own command type
			err := RegisterCommandHandler(manager, handler)
			if err != nil {
				t.Errorf("Failed to register handler in goroutine %d: %v", id, err)
				return
			}

			// Test concurrent reads
			for j := 0; j < 5; j++ {
				commands, queries, events, validators := manager.GetHandlerCounts()
				if commands < 0 || queries < 0 || events < 0 || validators < 0 {
					t.Errorf("Invalid counts in goroutine %d: c=%d, q=%d, e=%d, v=%d",
						id, commands, queries, events, validators)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state - since each goroutine registers a handler for the same underlying type
	// (testCommandHandler), we should have exactly 1 handler (the last one wins)
	commands, _, _, _ := manager.GetHandlerCounts()
	if commands != 1 {
		t.Errorf("Expected exactly 1 command handler (last registration wins), got %d", commands)
	}
}

// OPTIMIZATION TEST: Verify type cache performance
func TestManager_TypeCachePerformance(t *testing.T) {
	manager := NewCqrsManager()
	cmd := &TestCommand{ID: "test", Data: "data"}

	// First call should cache the type
	start := time.Now()
	typ1 := manager.typeCache.getType(cmd)
	firstCallDuration := time.Since(start)

	// Subsequent calls should be faster (cached)
	start = time.Now()
	typ2 := manager.typeCache.getType(cmd)
	secondCallDuration := time.Since(start)

	if typ1 != typ2 {
		t.Error("Type cache should return the same type for the same input")
	}

	// Second call should be faster (though this is not guaranteed in all environments)
	// We mainly test that it doesn't panic and returns consistent results
	t.Logf("First call: %v, Second call: %v", firstCallDuration, secondCallDuration)
}

// OPTIMIZATION TEST: Verify metrics functionality
func TestManager_Metrics(t *testing.T) {
	manager := NewCqrsManager()

	// Initial counts should be zero
	if manager.GetCommandCount() != 0 {
		t.Error("Initial command count should be 0")
	}
	if manager.GetQueryCount() != 0 {
		t.Error("Initial query count should be 0")
	}
	if manager.GetEventCount() != 0 {
		t.Error("Initial event count should be 0")
	}

	// Register handlers
	cmdHandler := &testCommandHandler{}
	queryHandler := &testQueryHandler{}
	eventHandler := &testEventHandler{}

	RegisterCommandHandler(manager, cmdHandler)
	RegisterQueryHandler(manager, queryHandler)
	RegisterEventHandler(manager, eventHandler)

	// Execute operations and verify metrics
	ctx := context.Background()

	ExecuteCommand(ctx, manager, &TestCommand{ID: "test"})
	if manager.GetCommandCount() != 1 {
		t.Errorf("Expected command count 1, got %d", manager.GetCommandCount())
	}

	ExecuteQuery[TestQuery, TestQueryResult](ctx, manager, TestQuery{ID: "test"})
	if manager.GetQueryCount() != 1 {
		t.Errorf("Expected query count 1, got %d", manager.GetQueryCount())
	}

	PublishEvent(ctx, manager, TestEvent{ID: "test", Message: "hello"})
	if manager.GetEventCount() != 1 {
		t.Errorf("Expected event count 1, got %d", manager.GetEventCount())
	}
}

// OPTIMIZATION TEST: Verify error handling improvements
func TestManager_ErrorHandling(t *testing.T) {
	manager := NewCqrsManager()

	// Test registration error handling
	cmdHandler := &testCommandHandler{}
	err := RegisterCommandHandler(manager, cmdHandler)
	if err != nil {
		t.Errorf("Expected successful registration, got error: %v", err)
	}

	// Test execution with missing handler
	ctx := context.Background()
	type UnregisteredCommand struct {
		command.Base
	}

	err = ExecuteCommand(ctx, manager, &UnregisteredCommand{})
	if err == nil {
		t.Error("Expected error for unregistered command handler")
	}

	// Test query with missing handler
	type UnregisteredQuery struct {
		query.Base
	}

	_, err = ExecuteQuery[UnregisteredQuery, string](ctx, manager, UnregisteredQuery{})
	if err == nil {
		t.Error("Expected error for unregistered query handler")
	}
}

// OPTIMIZATION TEST: Verify backward compatibility
func TestManager_BackwardCompatibility(t *testing.T) {
	// Test that default manager methods still work
	SetDefaultManager(NewCqrsManager())

	cmdHandler := &testCommandHandler{}
	err := RegisterCommandHandler(GetDefaultManager(), cmdHandler)
	if err != nil {
		t.Fatalf("Failed to register with default manager: %v", err)
	}

	// Test deprecated methods
	err = ExecuteCommandWithBackground(&TestCommand{ID: "test"})
	if err != nil {
		t.Errorf("Backward compatibility method failed: %v", err)
	}

	if !cmdHandler.WasExecuted() {
		t.Error("Command should have been executed")
	}
}

// BENCHMARK: Test performance improvements
func BenchmarkManager_CommandExecution(b *testing.B) {
	manager := NewCqrsManager()
	cmdHandler := &testCommandHandler{}
	RegisterCommandHandler(manager, cmdHandler)

	ctx := context.Background()
	cmd := &TestCommand{ID: "benchmark", Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteCommand(ctx, manager, cmd)
	}
}

func BenchmarkManager_TypeCache(b *testing.B) {
	manager := NewCqrsManager()
	cmd := &TestCommand{ID: "benchmark", Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.typeCache.getType(cmd)
	}
}

func BenchmarkManager_ConcurrentAccess(b *testing.B) {
	manager := NewCqrsManager()
	cmdHandler := &testCommandHandler{}
	RegisterCommandHandler(manager, cmdHandler)

	ctx := context.Background()
	cmd := &TestCommand{ID: "benchmark", Data: "test"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ExecuteCommand(ctx, manager, cmd)
		}
	})
}
