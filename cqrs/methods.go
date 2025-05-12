package cqrs

import (
	"fmt"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

func ExecuteCommand[T any](cmd T) error {
	typ := reflect.TypeOf(cmd)

	// Run validators on the command
	mgr.mu.RLock()
	validators := mgr.validators[typ]
	mgr.mu.RUnlock()

	for _, v := range validators {
		val, ok := v.(validator.IValidatorHandler[T])
		if !ok {
			return fmt.Errorf("validator type mismatch for %T", cmd)
		}
		if err := val.Validate(cmd); err != nil {
			return err
		}
	}

	// Run command if validators pass
	mgr.mu.RLock()
	handler, ok := mgr.commandHandlers[typ]
	mgr.mu.RUnlock()
	if !ok {
		return fmt.Errorf("handler not found for type %v", typ)
	}
	typedHandler, ok := handler.(command.ICommandHandler[T])
	if !ok {
		return fmt.Errorf("handler type mismatch for %v", typ)
	}

	// Pass the command to the handler
	return typedHandler.Handle(cmd)
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
