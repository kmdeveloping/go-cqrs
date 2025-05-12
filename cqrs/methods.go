// filepath: /Volumes/ExternalX1/Source/GolandProjects/go-cqrs/cqrs/methods.go
package cqrs

import (
	"fmt"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
)

// ExecuteCommand handles execution of pointer command types to support command state mutations
func ExecuteCommand[T any](cmd *T) error {
	typ := reflect.TypeOf(cmd)

	// Run command handler
	mgr.mu.RLock()
	handler, ok := mgr.commandHandlers[typ]
	mgr.mu.RUnlock()
	if !ok {
		return fmt.Errorf("handler not found for type %v", typ)
	}

	// Since we're accepting a pointer already, we can directly use it
	h, ok := handler.(interface{ Handle(*T) error })
	if !ok {
		return fmt.Errorf("handler type mismatch for %v", typ)
	}

	return h.Handle(cmd)
}

func ExecuteQuery[T query.IQuery, R any](qry T) (R, error) {
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

	return typedHandler.Handle(qry)
}

func PublishEvent[T event.IEvent](e T) error {
	typ := reflect.TypeOf(e)

	mgr.mu.RLock()
	handlerList := mgr.eventHandlers[typ]
	mgr.mu.RUnlock()

	for _, h := range handlerList {
		typedHandler, ok := h.(event.IEventHandler[T])
		if !ok {
			return fmt.Errorf("event handler type mismatch for %T", e)
		}

		if err := typedHandler.Handle(e); err != nil {
			return err
		}
	}

	return nil
}
