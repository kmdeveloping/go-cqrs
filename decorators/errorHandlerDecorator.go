package decorators

import (
	"context"
)

func ErrorHandlerDecorator() HandlerDecorator {
	return func(next AnyHandler) AnyHandler {
		return AnyHandlerFunc(func(ctx context.Context, message any) (any, error) {
			t, err := next.Handle(ctx, message)
			if err != nil {
				// do some handling stuff here
				return nil, err
			}

			return t, nil
		})
	}
}
