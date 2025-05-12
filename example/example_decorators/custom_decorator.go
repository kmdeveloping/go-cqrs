package example_decorators

import (
	"context"
	"log"

	"github.com/kmdeveloping/go-cqrs/decorators"
)

func ErrorHandlerDecorator() decorators.HandlerDecorator {
	return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
		return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
			log.Printf("Handling message: %T", message)
			return next.Handle(ctx, message)
		})
	}
}
