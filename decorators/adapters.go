package decorators

import (
	"context"
	"fmt"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
	"github.com/kmdeveloping/go-cqrs/validator"
)

func WithDecorators(base IHandlerDecorator, d ...HandlerDecorator) IHandlerDecorator {
	wrapped := base
	for i := len(d) - 1; i >= 0; i-- {
		wrapped = d[i](wrapped)
	}

	return wrapped
}

func WrapCommandHandler[T command.ICommand](h command.ICommandHandler[T]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		cmd, ok := message.(*T)
		if !ok {
			return nil, fmt.Errorf("invalid command type: %T", message)
		}
		err := h.Handle(ctx, cmd)
		return nil, err
	})
}

func UnwrapAsCommandHandler[T command.ICommand](h IHandlerDecorator) (command.ICommandHandler[T], bool) {
	return commandHandlerFunc[T](func(ctx context.Context, cmd *T) error {
		_, err := h.Handle(ctx, cmd)
		return err
	}), true
}

func WrapQueryHandler[T query.IQuery, R any](h query.IQueryHandler[T, R]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		q, ok := message.(T)
		if !ok {
			return nil, fmt.Errorf("invalid query type: %T", message)
		}
		return h.Handle(ctx, q)
	})
}

func UnwrapAsQueryHandler[T query.IQuery, R any](h IHandlerDecorator) (query.IQueryHandler[T, R], bool) {
	return queryHandlerFunc[T, R](func(ctx context.Context, query T) (R, error) {
		res, err := h.Handle(ctx, query)
		if err != nil {
			var zero R
			return zero, err
		}
		r, ok := res.(R)
		if !ok {
			var zero R
			return zero, fmt.Errorf("invalid query result type")
		}
		return r, nil
	}), true
}

func WrapEventHandler[T event.IEvent](h event.IEventHandler[T]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		e, ok := message.(T)
		if !ok {
			return nil, fmt.Errorf("invalid event type: %T", message)
		}
		return nil, h.Handle(ctx, e)
	})
}

func UnwrapAsEventHandler[T event.IEvent](h IHandlerDecorator) (event.IEventHandler[T], bool) {
	return eventHandlerFunc[T](func(ctx context.Context, e T) error {
		_, err := h.Handle(ctx, e)
		return err
	}), true
}

// Validator wrapper functions
func WrapValidator[T any](v validator.IValidatorHandler[T]) IHandlerDecorator {
	return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
		cmd, ok := message.(*T)
		if !ok {
			return nil, fmt.Errorf("invalid command type for validator: %T", message)
		}
		err := v.Validate(ctx, cmd)
		return nil, err
	})
}

func UnwrapAsValidator[T any](h IHandlerDecorator) (validator.IValidatorHandler[T], bool) {
	return validatorFunc[T](func(ctx context.Context, cmd *T) error {
		_, err := h.Handle(ctx, cmd)
		return err
	}), true
}
