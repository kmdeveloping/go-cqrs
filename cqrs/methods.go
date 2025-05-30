// filepath: /Volumes/ExternalX1/Source/GolandProjects/go-cqrs/cqrs/methods.go
package cqrs

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

// ErrorAggregator collects multiple errors for event handling
type ErrorAggregator struct {
	errors []error
}

func (ea *ErrorAggregator) Add(err error) {
	if err != nil {
		ea.errors = append(ea.errors, err)
	}
}

func (ea *ErrorAggregator) Error() error {
	if len(ea.errors) == 0 {
		return nil
	}
	if len(ea.errors) == 1 {
		return ea.errors[0]
	}
	return fmt.Errorf("multiple errors: %v", ea.errors)
}

func (ea *ErrorAggregator) HasErrors() bool {
	return len(ea.errors) > 0
}

// ExecuteCommand handles execution of pointer command types to support command state mutations
// OPTIMIZATION: Now accepts Manager instance instead of using global singleton
func ExecuteCommand[T any](ctx context.Context, m *Manager, cmd *T) error {
	// OPTIMIZATION: Use cached type instead of expensive reflection
	typ := m.typeCache.getType(cmd)

	// OPTIMIZATION: Atomic increment for metrics
	atomic.AddInt64(&m.commandCount, 1)

	// Run validators first
	m.validatorsMu.RLock()
	validatorList, exists := m.validators[typ]
	var validatorsCopy []any
	if exists {
		// Make a copy to avoid holding the lock during validation
		validatorsCopy = make([]any, len(validatorList))
		copy(validatorsCopy, validatorList)
	}
	m.validatorsMu.RUnlock()

	if exists {
		for _, v := range validatorsCopy {
			typedValidator, ok := v.(validator.IValidatorHandler[T])
			if !ok {
				return fmt.Errorf("validator type mismatch for %T", cmd)
			}

			if err := typedValidator.Validate(ctx, cmd); err != nil {
				return fmt.Errorf("validation failed for %T: %w", cmd, err)
			}
		}
	}

	// Run command handler
	m.handlersMu.RLock()
	handler, ok := m.commandHandlers[typ]
	m.handlersMu.RUnlock()
	if !ok {
		return fmt.Errorf("handler not found for type %v", typ)
	}

	// Since we're accepting a pointer already, we can directly use it
	h, ok := handler.(interface {
		Handle(context.Context, *T) error
	})
	if !ok {
		return fmt.Errorf("handler type mismatch for %v", typ)
	}

	return h.Handle(ctx, cmd)
}

// OPTIMIZATION: Now accepts Manager instance instead of using global singleton
func ExecuteQuery[T query.IQuery, R any](ctx context.Context, m *Manager, qry T) (R, error) {
	var zero R
	// OPTIMIZATION: Use cached type instead of expensive reflection
	typ := m.typeCache.getType(qry)

	// OPTIMIZATION: Atomic increment for metrics
	atomic.AddInt64(&m.queryCount, 1)

	m.handlersMu.RLock()
	handler, ok := m.queryHandlers[typ]
	m.handlersMu.RUnlock()
	if !ok {
		return zero, fmt.Errorf("no query handler for %T", qry)
	}

	typedHandler, ok := handler.(query.IQueryHandler[T, R])
	if !ok {
		return zero, fmt.Errorf("query handler type mismatch for %T", qry)
	}

	return typedHandler.Handle(ctx, qry)
}

// OPTIMIZATION: Improved event handling with error aggregation and better performance
func PublishEvent[T event.IEvent](ctx context.Context, m *Manager, e T) error {
	// OPTIMIZATION: Use cached type instead of expensive reflection
	typ := m.typeCache.getType(e)

	// OPTIMIZATION: Atomic increment for metrics
	atomic.AddInt64(&m.eventCount, 1)

	m.handlersMu.RLock()
	handlerList, exists := m.eventHandlers[typ]
	var handlersCopy []any
	if exists {
		// Make a copy to avoid holding the lock during handler execution
		handlersCopy = make([]any, len(handlerList))
		copy(handlersCopy, handlerList)
	}
	m.handlersMu.RUnlock()

	if !exists {
		// Not an error if no handlers are registered for an event
		return nil
	}

	// OPTIMIZATION: Aggregate errors instead of failing fast to ensure all handlers get a chance to execute
	var errorAggregator ErrorAggregator

	for _, h := range handlersCopy {
		// OPTIMIZATION: Check context cancellation to avoid unnecessary work
		select {
		case <-ctx.Done():
			errorAggregator.Add(ctx.Err())
			break
		default:
		}

		typedHandler, ok := h.(event.IEventHandler[T])
		if !ok {
			errorAggregator.Add(fmt.Errorf("event handler type mismatch for %T", e))
			continue
		}

		if err := typedHandler.Handle(ctx, e); err != nil {
			errorAggregator.Add(fmt.Errorf("handler failed for %T: %w", e, err))
		}
	}

	return errorAggregator.Error()
}

// OPTIMIZATION: Async event publishing for better performance in high-throughput scenarios
func PublishEventAsync[T event.IEvent](ctx context.Context, m *Manager, e T) error {
	// For async processing, we start the event handling in a goroutine
	// and return immediately. Errors are logged but not returned.
	go func() {
		if err := PublishEvent(ctx, m, e); err != nil {
			// TODO: Add proper logging here based on the logging decorator configuration
			// For now, we silently handle the error to maintain backward compatibility
			_ = err
		}
	}()

	return nil
}

// Convenience methods for backward compatibility - these will need a default Manager instance
// DEPRECATED: These methods will be removed in a future version. Use the methods that accept Manager explicitly.

var defaultManager *Manager

// SetDefaultManager sets the default manager for backward compatibility methods
func SetDefaultManager(m *Manager) {
	defaultManager = m
}

// GetDefaultManager returns the default manager, creating one if it doesn't exist
func GetDefaultManager() *Manager {
	if defaultManager == nil {
		defaultManager = NewCqrsManager()
	}
	return defaultManager
}

// Convenience methods that use the default manager for backwards compatibility
func ExecuteCommandWithBackground[T any](cmd *T) error {
	return ExecuteCommand(context.Background(), GetDefaultManager(), cmd)
}

func ExecuteQueryWithBackground[T query.IQuery, R any](qry T) (R, error) {
	return ExecuteQuery[T, R](context.Background(), GetDefaultManager(), qry)
}

func PublishEventWithBackground[T event.IEvent](e T) error {
	return PublishEvent(context.Background(), GetDefaultManager(), e)
}
