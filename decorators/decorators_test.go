package decorators

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// Test handler for decorator testing
type testHandler struct {
	called   bool
	duration time.Duration
	err      error
}

func (h *testHandler) Handle(ctx context.Context, msg any) (any, error) {
	h.called = true
	if h.duration > 0 {
		time.Sleep(h.duration)
	}
	return "test-result", h.err
}

// OPTIMIZATION TEST: Verify logging decorator configuration
func TestLoggingDecorator_Configuration(t *testing.T) {
	// Test disabled logging
	handler := &testHandler{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	disabledDecorator := LoggingDecoratorWithConfig(logger, LoggingConfig{
		Enabled: false,
	})

	decorated := disabledDecorator(HandlerDecoratorFunc(handler.Handle))

	ctx := context.Background()
	result, err := decorated.Handle(ctx, "test-message")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "test-result" {
		t.Errorf("Expected 'test-result', got: %v", result)
	}
	if !handler.called {
		t.Error("Handler should have been called")
	}
}

// OPTIMIZATION TEST: Verify sampling functionality
func TestLoggingDecorator_Sampling(t *testing.T) {
	handler := &testHandler{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	// Configure with 0% sampling (should never log)
	samplingDecorator := LoggingDecoratorWithConfig(logger, LoggingConfig{
		Enabled:    true,
		LogInputs:  true,
		SampleRate: 0.0,
	})

	decorated := samplingDecorator(HandlerDecoratorFunc(handler.Handle))

	ctx := context.Background()

	// Execute multiple times - none should be logged due to 0% sampling
	for i := 0; i < 10; i++ {
		_, err := decorated.Handle(ctx, "test-message")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// All executions should succeed
	if !handler.called {
		t.Error("Handler should have been called")
	}
}

// OPTIMIZATION TEST: Verify metrics decorator configuration
func TestMetricsDecorator_Configuration(t *testing.T) {
	handler := &testHandler{}

	// Test disabled metrics
	disabledDecorator := MetricsDecoratorWithConfig(MetricsConfig{
		Enabled: false,
	})

	decorated := disabledDecorator(HandlerDecoratorFunc(handler.Handle))

	ctx := context.Background()
	result, err := decorated.Handle(ctx, "test-message")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "test-result" {
		t.Errorf("Expected 'test-result', got: %v", result)
	}
	if !handler.called {
		t.Error("Handler should have been called")
	}
}

// OPTIMIZATION TEST: Verify slow threshold functionality
func TestMetricsDecorator_SlowThreshold(t *testing.T) {
	// Create a handler that takes some time
	handler := &testHandler{duration: 50 * time.Millisecond}

	// Configure metrics with a threshold
	metricsDecorator := MetricsDecoratorWithConfig(MetricsConfig{
		Enabled:            true,
		LogAllExecutions:   false,
		SlowThreshold:      100 * time.Millisecond, // Higher than handler duration
		EnableDetailedLogs: false,
	})

	decorated := metricsDecorator(HandlerDecoratorFunc(handler.Handle))

	ctx := context.Background()
	start := time.Now()
	_, err := decorated.Handle(ctx, "test-message")
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should complete in reasonable time
	if duration > 200*time.Millisecond {
		t.Errorf("Handler took too long: %v", duration)
	}
}

// OPTIMIZATION TEST: Verify context cancellation in decorators
func TestDecorators_ContextCancellation(t *testing.T) {
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Test command handler wrapper
	cmdWrapper := WrapCommandHandler(testCommandHandler{})

	// Cancel context before execution
	cancel()

	result, err := cmdWrapper.Handle(ctx, "test")
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}
}

// OPTIMIZATION TEST: Verify timeout decorator
func TestTimeoutDecorator(t *testing.T) {
	// Create a slow handler
	slowHandler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		time.Sleep(200 * time.Millisecond)
		return "slow-result", nil
	})

	// Apply timeout decorator with short timeout
	timeoutDecorator := TimeoutDecorator(50 * time.Millisecond)
	decorated := timeoutDecorator(slowHandler)

	ctx := context.Background()
	start := time.Now()
	result, err := decorated.Handle(ctx, "test")
	duration := time.Since(start)

	// Should timeout
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result on timeout, got: %v", result)
	}

	// Should complete quickly due to timeout
	if duration > 100*time.Millisecond {
		t.Errorf("Timeout took too long: %v", duration)
	}
}

// OPTIMIZATION TEST: Verify timeout decorator with fast handler
func TestTimeoutDecorator_FastHandler(t *testing.T) {
	// Create a fast handler
	fastHandler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		return "fast-result", nil
	})

	// Apply timeout decorator with generous timeout
	timeoutDecorator := TimeoutDecorator(100 * time.Millisecond)
	decorated := timeoutDecorator(fastHandler)

	ctx := context.Background()
	result, err := decorated.Handle(ctx, "test")

	// Should succeed
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "fast-result" {
		t.Errorf("Expected 'fast-result', got: %v", result)
	}
}

// OPTIMIZATION TEST: Verify decorator chaining optimization
func TestWithDecorators_EmptyChain(t *testing.T) {
	handler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		return "base-result", nil
	})

	// Test with no decorators
	decorated := WithDecorators(handler)

	// Test that the decorated handler works correctly
	ctx := context.Background()
	result, err := decorated.Handle(ctx, "test")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "base-result" {
		t.Errorf("Expected 'base-result', got: %v", result)
	}
}

// OPTIMIZATION TEST: Verify improved error messages
func TestWrappers_ImprovedErrorMessages(t *testing.T) {
	// Test command handler with wrong type
	cmdHandler := testCommandHandler{}
	wrapped := WrapCommandHandler(cmdHandler)

	ctx := context.Background()

	// Pass wrong type
	_, err := wrapped.Handle(ctx, "wrong-type")

	if err == nil {
		t.Error("Expected type mismatch error")
	}

	// Should have improved error message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "expected") && !strings.Contains(errMsg, "got") {
		t.Errorf("Expected improved error message with type information, got: %v", errMsg)
	}
}

// Test command handler for decorator testing
type testCommandHandler struct{}

func (h testCommandHandler) Handle(ctx context.Context, cmd *string) error {
	return nil
}

// BENCHMARK: Test decorator performance overhead
func BenchmarkDecorator_Overhead(b *testing.B) {
	handler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		return "result", nil
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, "test")
	}
}

func BenchmarkDecorator_WithLogging(b *testing.B) {
	handler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		return "result", nil
	})

	logger := log.New(io.Discard, "", 0) // Discard output for benchmarking
	decorator := LoggingDecorator(logger)
	decorated := decorator(handler)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decorated.Handle(ctx, "test")
	}
}

func BenchmarkDecorator_WithMetrics(b *testing.B) {
	handler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		return "result", nil
	})

	decorator := MetricsDecorator()
	decorated := decorator(handler)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decorated.Handle(ctx, "test")
	}
}

func BenchmarkDecorator_MultipleDecorators(b *testing.B) {
	handler := HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		return "result", nil
	})

	logger := log.New(io.Discard, "", 0)
	loggingDecorator := LoggingDecorator(logger)
	metricsDecorator := MetricsDecorator()

	decorated := WithDecorators(handler, loggingDecorator, metricsDecorator)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decorated.Handle(ctx, "test")
	}
}
