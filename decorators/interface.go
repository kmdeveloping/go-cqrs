package decorators

import "context"

type AnyHandler interface {
	Handle(ctx context.Context, msg any) (any, error)
}

type HandlerDecorator func(next AnyHandler) AnyHandler
