package decorators

import (
	"context"
	"log"
)

func LoggingDecorator(logger *log.Logger) HandlerDecorator {
	return func(next AnyHandler) AnyHandler {
		return AnyHandlerFunc(func(ctx context.Context, message any) (any, error) {
			logger.Printf("[Handler] %T => %+v", message, message)
			return next.Handle(ctx, message)
		})
	}
}
