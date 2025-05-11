package decorators

import (
	"context"
	"log"
	"time"
)

func MetricsDecorator() HandlerDecorator {
	return func(next IHandlerDecorator) IHandlerDecorator {
		return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
			start := time.Now()
			log.Printf("Execution started @ %s\n", start.Format(time.RFC3339Nano))

			t, err := next.Handle(ctx, message)

			stop := time.Now()
			log.Printf("Execution completed @ %s\t total time: %s\n", stop.Format(time.RFC3339Nano), stop.Sub(start).String())

			return t, err
		})
	}
}
