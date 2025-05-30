package decorators

import (
	"context"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
)

type HandlerDecoratorFunc func(ctx context.Context, msg any) (any, error)

func (f HandlerDecoratorFunc) Handle(ctx context.Context, msg any) (any, error) { return f(ctx, msg) }

type commandHandlerFunc[T any] func(context.Context, *T) error

func (f commandHandlerFunc[T]) Handle(ctx context.Context, cmd *T) error { return f(ctx, cmd) }

type queryHandlerFunc[T query.IQuery, R any] func(context.Context, T) (R, error)

func (f queryHandlerFunc[T, R]) Handle(ctx context.Context, q T) (R, error) { return f(ctx, q) }

type eventHandlerFunc[T event.IEvent] func(context.Context, T) error

func (f eventHandlerFunc[T]) Handle(ctx context.Context, e T) error { return f(ctx, e) }

type validatorFunc[T any] func(context.Context, *T) error

func (f validatorFunc[T]) Validate(ctx context.Context, cmd *T) error { return f(ctx, cmd) }
