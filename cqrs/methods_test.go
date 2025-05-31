package cqrs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test event handlers for error aggregation testing
type successEventHandler struct {
	handled bool
	mu      sync.Mutex
}

func (h *successEventHandler) Handle(ctx context.Context, e TestEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handled = true
	return nil
}

func (h *successEventHandler) WasHandled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.handled
}

type errorEventHandler struct {
	handled bool
	mu      sync.Mutex
	err     error
}

func (h *errorEventHandler) Handle(ctx context.Context, e TestEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handled = true
	return h.err
}

func (h *errorEventHandler) WasHandled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.handled
}

// OPTIMIZATION TEST: Verify error aggregation in event publishing
func TestPublishEvent_ErrorAggregation(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	ctx := context.Background()

	// Register multiple handlers - some succeed, some fail
	successHandler1 := &successEventHandler{}
	successHandler2 := &successEventHandler{}
	errorHandler1 := &errorEventHandler{err: errors.New("error 1")}
	errorHandler2 := &errorEventHandler{err: errors.New("error 2")}

	RegisterEventHandler(successHandler1)
	RegisterEventHandler(errorHandler1)
	RegisterEventHandler(successHandler2)
	RegisterEventHandler(errorHandler2)

	// Publish event
	err := PublishEvent(ctx, TestEvent{ID: "test", Message: "hello"})

	// Verify all handlers were called
	if !successHandler1.WasHandled() {
		t.Error("Success handler 1 should have been called")
	}
	if !successHandler2.WasHandled() {
		t.Error("Success handler 2 should have been called")
	}
	if !errorHandler1.WasHandled() {
		t.Error("Error handler 1 should have been called")
	}
	if !errorHandler2.WasHandled() {
		t.Error("Error handler 2 should have been called")
	}

	// Verify error aggregation
	if err == nil {
		t.Error("Expected aggregated error, got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "error 1") || !strings.Contains(errMsg, "error 2") {
			t.Errorf("Expected aggregated error to contain both errors, got: %v", errMsg)
		}
	}
}

// OPTIMIZATION TEST: Verify single error handling
func TestPublishEvent_SingleError(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	ctx := context.Background()

	// Register one error handler
	errorHandler := &errorEventHandler{err: errors.New("single error")}
	RegisterEventHandler(errorHandler)

	err := PublishEvent(ctx, TestEvent{ID: "test", Message: "hello"})

	if err == nil {
		t.Error("Expected error, got nil")
	}

	// The error is now wrapped with additional context, so check for the original error
	if !strings.Contains(err.Error(), "single error") {
		t.Errorf("Expected error containing 'single error', got: %v", err)
	}
}

// OPTIMIZATION TEST: Verify no error when all handlers succeed
func TestPublishEvent_NoError(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	ctx := context.Background()

	// Register success handlers
	successHandler1 := &successEventHandler{}
	successHandler2 := &successEventHandler{}
	RegisterEventHandler(successHandler1)
	RegisterEventHandler(successHandler2)

	err := PublishEvent(ctx, TestEvent{ID: "test", Message: "hello"})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// OPTIMIZATION TEST: Verify context cancellation handling
func TestPublishEvent_ContextCancellation(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Register a slow handler that should be interrupted
	slowHandler := &testEventHandler{}
	RegisterEventHandler(slowHandler)

	// Cancel the context before publishing
	cancel()

	err := PublishEvent(ctx, TestEvent{ID: "test", Message: "hello"})

	// Should return context cancellation error
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

// OPTIMIZATION TEST: Verify async event publishing
func TestPublishEventAsync_NonBlocking(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	ctx := context.Background()

	// Register a slow handler
	slowHandler := &testEventHandler{}
	RegisterEventHandler(slowHandler)

	// Async publishing should return immediately
	start := time.Now()
	err := PublishEventAsync(ctx, TestEvent{ID: "test", Message: "hello"})
	duration := time.Since(start)

	// Should return immediately without error
	if err != nil {
		t.Errorf("Async publish should not return error, got: %v", err)
	}

	// Should be very fast (less than 10ms)
	if duration > 10*time.Millisecond {
		t.Errorf("Async publish took too long: %v", duration)
	}

	// Give some time for the async handler to complete
	time.Sleep(50 * time.Millisecond)

	// Handler should have been called
	if !slowHandler.WasHandled() {
		t.Error("Async handler should have been called")
	}
}

// OPTIMIZATION TEST: Verify ErrorAggregator functionality
func TestErrorAggregator(t *testing.T) {
	// Test empty aggregator
	var ea ErrorAggregator
	if ea.HasErrors() {
		t.Error("Empty aggregator should not have errors")
	}
	if ea.Error() != nil {
		t.Error("Empty aggregator should return nil error")
	}

	// Test single error
	ea.Add(errors.New("error 1"))
	if !ea.HasErrors() {
		t.Error("Aggregator with error should have errors")
	}
	if ea.Error().Error() != "error 1" {
		t.Errorf("Single error should be returned as-is, got: %v", ea.Error())
	}

	// Test multiple errors
	ea.Add(errors.New("error 2"))
	ea.Add(errors.New("error 3"))

	err := ea.Error()
	if err == nil {
		t.Error("Multiple errors should return non-nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "multiple errors") {
		t.Errorf("Multiple errors should be indicated, got: %v", errMsg)
	}

	// Test adding nil error (should be ignored)
	initialErrorCount := len(ea.errors)
	ea.Add(nil)
	if len(ea.errors) != initialErrorCount {
		t.Error("Adding nil error should not increase error count")
	}
}

// OPTIMIZATION TEST: Verify no handlers scenario
func TestPublishEvent_NoHandlers(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	ctx := context.Background()

	// Publish event without any registered handlers
	err := PublishEvent(ctx, TestEvent{ID: "test", Message: "hello"})

	// Should not return an error
	if err != nil {
		t.Errorf("Publishing event with no handlers should not error, got: %v", err)
	}
}

// OPTIMIZATION TEST: Verify command validation with type cache
func TestExecuteCommand_WithValidation(t *testing.T) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	ctx := context.Background()

	// Register validator and handler
	validator := &testValidator{}
	handler := &testCommandHandler{}

	RegisterValidator(validator)
	RegisterCommandHandler(handler)

	// Execute command
	err := ExecuteCommand(ctx, &TestCommand{ID: "test", Data: "data"})

	if err != nil {
		t.Errorf("Command execution should succeed, got: %v", err)
	}

	// Verify validator was called
	if !validator.WasValidated() {
		t.Error("Validator should have been called")
	}

	// Verify handler was called
	if !handler.WasExecuted() {
		t.Error("Handler should have been called")
	}
}

// BENCHMARK: Test event publishing performance
func BenchmarkPublishEvent_SingleHandler(b *testing.B) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)
	handler := &testEventHandler{}
	RegisterEventHandler(handler)

	ctx := context.Background()
	event := TestEvent{ID: "benchmark", Message: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PublishEvent(ctx, event)
	}
}

func BenchmarkPublishEvent_MultipleHandlers(b *testing.B) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)

	// Register multiple handlers
	for i := 0; i < 10; i++ {
		handler := &testEventHandler{}
		RegisterEventHandler(handler)
	}

	ctx := context.Background()
	event := TestEvent{ID: "benchmark", Message: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PublishEvent(ctx, event)
	}
}

func BenchmarkPublishEvent_ErrorAggregation(b *testing.B) {
	cleanup := setupTestManager()
	defer cleanup()

	manager := NewCqrsManager()
	SetManager(manager)

	// Register mix of success and error handlers
	for i := 0; i < 5; i++ {
		RegisterEventHandler(&successEventHandler{})
		RegisterEventHandler(&errorEventHandler{err: fmt.Errorf("error %d", i)})
	}

	ctx := context.Background()
	event := TestEvent{ID: "benchmark", Message: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PublishEvent(ctx, event)
	}
}
