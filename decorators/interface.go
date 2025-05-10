package decorators

import "context"

type IHandlerDecorator interface {
	Handle(ctx context.Context, msg any) (any, error)
}

type HandlerDecorator func(next IHandlerDecorator) IHandlerDecorator
