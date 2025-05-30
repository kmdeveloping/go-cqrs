// filepath: /Volumes/ExternalX1/Source/GolandProjects/go-cqrs/cqrs/methods.go
package cqrs

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

// ExecuteCommand handles execution of pointer command types to support command state mutations
func ExecuteCommand[T any](ctx context.Context, cmd *T) error {
	typ := reflect.TypeOf(cmd)

	// Run validators first
	mgr.mu.RLock()
	validatorList := mgr.validators[typ]
	mgr.mu.RUnlock()

	for _, v := range validatorList {
		typedValidator, ok := v.(validator.IValidatorHandler[T])
		if !ok {
			return fmt.Errorf("validator type mismatch for %T", cmd)
		}

		if err := typedValidator.Validate(ctx, cmd); err != nil {
			return fmt.Errorf("validation failed for %T: %w", cmd, err)
		}
	}

	// Run command handler
	mgr.mu.RLock()
	handler, ok := mgr.commandHandlers[typ]
	mgr.mu.RUnlock()
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

func ExecuteQuery[T query.IQuery, R any](ctx context.Context, qry T) (R, error) {
	var zero R
	typ := reflect.TypeOf(qry)

	mgr.mu.RLock()
	handler, ok := mgr.queryHandlers[typ]
	mgr.mu.RUnlock()
	if !ok {
		return zero, fmt.Errorf("no query handler for %T", qry)
	}

	typedHandler, ok := handler.(query.IQueryHandler[T, R])
	if !ok {
		return zero, fmt.Errorf("query handler type mismatch for %T", qry)
	}

	return typedHandler.Handle(ctx, qry)
}

func PublishEvent[T event.IEvent](ctx context.Context, e T) error {
	typ := reflect.TypeOf(e)

	mgr.mu.RLock()
	handlerList := mgr.eventHandlers[typ]
	mgr.mu.RUnlock()

	for _, h := range handlerList {
		typedHandler, ok := h.(event.IEventHandler[T])
		if !ok {
			return fmt.Errorf("event handler type mismatch for %T", e)
		}

		if err := typedHandler.Handle(ctx, e); err != nil {
			return err
		}
	}

	return nil
}

// Convenience methods that use context.Background() for backwards compatibility
func ExecuteCommandWithBackground[T any](cmd *T) error {
	return ExecuteCommand(context.Background(), cmd)
}

func ExecuteQueryWithBackground[T query.IQuery, R any](qry T) (R, error) {
	return ExecuteQuery[T, R](context.Background(), qry)
}

func PublishEventWithBackground[T event.IEvent](e T) error {
	return PublishEvent(context.Background(), e)
}
