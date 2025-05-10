package decorators

import (
	"context"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
)

type HandlerDecoratorFunc func(ctx context.Context, msg any) (any, error)

func (f HandlerDecoratorFunc) Handle(ctx context.Context, msg any) (any, error) { return f(ctx, msg) }

type commandHandlerFunc[T command.ICommand] func(T) error

func (f commandHandlerFunc[T]) Handle(cmd T) error { return f(cmd) }

type queryHandlerFunc[T query.IQuery, R any] func(T) (R, error)

func (f queryHandlerFunc[T, R]) Handle(q T) (R, error) { return f(q) }

type eventHandlerFunc[T event.IEvent] func(T) error

func (f eventHandlerFunc[T]) Handle(e T) error { return f(e) }
