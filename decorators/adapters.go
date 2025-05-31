package decorators

import (
	"context"
	"fmt"
	"time"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

// OPTIMIZATION: Improved decorator chaining with better error handling
func WithDecorators(base IHandlerDecorator, d ...HandlerDecorator) IHandlerDecorator {
	if len(d) == 0 {
		return base
	}

	wrapped := base
	for i := len(d) - 1; i >= 0; i-- {
		wrapped = d[i](wrapped)
	}

	return wrapped
}

// OPTIMIZATION: Context-aware command handler wrapper
func WrapCommandHandler[T command.ICommand](h command.ICommandHandler[T]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		// OPTIMIZATION: Check context cancellation early
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		cmd, ok := message.(*T)
		if !ok {
			return nil, fmt.Errorf("invalid command type: expected *%T, got %T", new(T), message)
		}
		err := h.Handle(ctx, cmd)
		return nil, err
	})
}

// OPTIMIZATION: Improved type safety in unwrapping
func UnwrapAsCommandHandler[T command.ICommand](h IHandlerDecorator) (command.ICommandHandler[T], bool) {
	wrapper := commandHandlerFunc[T](func(ctx context.Context, cmd *T) error {
		_, err := h.Handle(ctx, cmd)
		return err
	})
	return wrapper, true
}

// OPTIMIZATION: Context-aware query handler wrapper
func WrapQueryHandler[T query.IQuery, R any](h query.IQueryHandler[T, R]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		// OPTIMIZATION: Check context cancellation early
		select {
		case <-ctx.Done():
			var zero R
			return zero, ctx.Err()
		default:
		}

		q, ok := message.(T)
		if !ok {
			var zero R
			return zero, fmt.Errorf("invalid query type: expected %T, got %T", *new(T), message)
		}
		return h.Handle(ctx, q)
	})
}

// OPTIMIZATION: Improved type safety and error handling
func UnwrapAsQueryHandler[T query.IQuery, R any](h IHandlerDecorator) (query.IQueryHandler[T, R], bool) {
	wrapper := queryHandlerFunc[T, R](func(ctx context.Context, query T) (R, error) {
		res, err := h.Handle(ctx, query)
		if err != nil {
			var zero R
			return zero, err
		}
		r, ok := res.(R)
		if !ok {
			var zero R
			return zero, fmt.Errorf("invalid query result type: expected %T, got %T", *new(R), res)
		}
		return r, nil
	})
	return wrapper, true
}

// OPTIMIZATION: Context-aware event handler wrapper
func WrapEventHandler[T event.IEvent](h event.IEventHandler[T]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		// OPTIMIZATION: Check context cancellation early
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		e, ok := message.(T)
		if !ok {
			return nil, fmt.Errorf("invalid event type: expected %T, got %T", *new(T), message)
		}
		return nil, h.Handle(ctx, e)
	})
}

// OPTIMIZATION: Improved error handling in event unwrapping
func UnwrapAsEventHandler[T event.IEvent](h IHandlerDecorator) (event.IEventHandler[T], bool) {
	wrapper := eventHandlerFunc[T](func(ctx context.Context, e T) error {
		_, err := h.Handle(ctx, e)
		return err
	})
	return wrapper, true
}

// Validator wrapper functions - OPTIMIZATION: Added context cancellation support
func WrapValidator[T any](v validator.IValidatorHandler[T]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		// OPTIMIZATION: Check context cancellation early
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		cmd, ok := message.(*T)
		if !ok {
			return nil, fmt.Errorf("invalid command type for validator: expected *%T, got %T", new(T), message)
		}
		err := v.Validate(ctx, cmd)
		return nil, err
	})
}

// OPTIMIZATION: Improved validator unwrapping
func UnwrapAsValidator[T any](h IHandlerDecorator) (validator.IValidatorHandler[T], bool) {
	wrapper := validatorFunc[T](func(ctx context.Context, cmd *T) error {
		_, err := h.Handle(ctx, cmd)
		return err
	})
	return wrapper, true
}

// OPTIMIZATION: New decorator for timeout handling
func TimeoutDecorator(timeout time.Duration) HandlerDecorator {
	return func(next IHandlerDecorator) IHandlerDecorator {
		return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
			if timeout <= 0 {
				return next.Handle(ctx, message)
			}

			// Create a context with timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Use a channel to handle the result
			type result struct {
				value any
				err   error
			}

			resultChan := make(chan result, 1)
			go func() {
				val, err := next.Handle(timeoutCtx, message)
				resultChan <- result{val, err}
			}()

			select {
			case res := <-resultChan:
				return res.value, res.err
			case <-timeoutCtx.Done():
				return nil, fmt.Errorf("handler timeout after %v for %T", timeout, message)
			}
		})
	}
}
