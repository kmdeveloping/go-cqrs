package cqrs

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/decorators"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
	"reflect"
)

func ExecuteCommand[T command.ICommand](bus *Manager, cmd T) error {
	typ := reflect.TypeOf(cmd)

	// Run validators on command
	bus.mu.RLock()
	validators := bus.validators[typ]
	bus.mu.RUnlock()

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
	bus.mu.RLock()
	handler, ok := bus.commandHandlers[typ]
	bus.mu.RUnlock()
	if !ok {
		return fmt.Errorf("handler not found for type %v", typ)
	}

	typedHandler, ok := handler.(command.ICommandHandler[T])
	if !ok {
		return fmt.Errorf("handler type mismatch for %v", typ)
	}

	/// order decorators last is first to be called
	/// call stack: metrics >> errorHandler >> logger >> handler method
	if bus.config.enableLoggingDecorator {
		typedHandler = decorators.UseLoggingDecorator(typedHandler)
	}

	if bus.config.enableErrorHandlerDecorator {
		typedHandler = decorators.UseErrorHandlerDecorator(typedHandler)
	}

	if bus.config.enableMetricsDecorator {
		typedHandler = decorators.UseExecutionTimeDecorator(typedHandler)
	}

	return typedHandler.Handle(cmd)
}

func ExecuteQuery[T query.IQuery, R any](bus *Manager, qry T) (R, error) {
	var zero R
	typ := reflect.TypeOf(qry)
	bus.mu.RLock()
	handler, ok := bus.queryHandlers[typ]
	bus.mu.RUnlock()
	if !ok {
		return zero, fmt.Errorf("no query handler for %T", qry)
	}

	typedHandler, ok := handler.(query.IQueryHandler[T, R])
	if !ok {
		return zero, fmt.Errorf("query handler type mismatch for %T", qry)
	}

	/// order decorators last is first to be called
	/// call stack: metrics >> errorHandler >> logger >> handler method
	/*if bus.config.enableLoggingDecorator {
		typedHandler = decorators.UseLoggingDecorator(typedHandler)
	}

	if bus.config.enableErrorHandlerDecorator {
		typedHandler = decorators.UseErrorHandlerDecorator(typedHandler)
	}

	if bus.config.enableMetricsDecorator {
		typedHandler = decorators.UseExecutionTimeDecorator(typedHandler)
	}*/

	return typedHandler.Handle(qry)
}

func PublishEvent[T event.IEvent](bus *Manager, e T) error {
	typ := reflect.TypeOf(e)

	bus.mu.RLock()
	handlerList := bus.eventHandlers[typ]
	bus.mu.RUnlock()

	for _, h := range handlerList {
		typedHandler, ok := h.(event.IEventHandler[T])
		if !ok {
			return fmt.Errorf("event handler type mismatch for %T", e)
		}

		/// order decorators last is first to be called
		/// call stack: metrics >> errorHandler >> logger >> handler method
		/*if bus.config.enableLoggingDecorator {
			typedHandler = decorators.UseLoggingDecorator(typedHandler)
		}

		if bus.config.enableErrorHandlerDecorator {
			typedHandler = decorators.UseErrorHandlerDecorator(typedHandler)
		}

		if bus.config.enableMetricsDecorator {
			typedHandler = decorators.UseExecutionTimeDecorator(typedHandler)
		}*/

		if err := typedHandler.Handle(e); err != nil {
			return err
		}
	}

	return nil
}
