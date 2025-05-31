package validator

import "context"

// IValidatorHandler validates commands of any type
// Commands are passed as pointers to validators for consistency
type IValidatorHandler[T any] interface {
	Validate(ctx context.Context, cmd *T) error
}
