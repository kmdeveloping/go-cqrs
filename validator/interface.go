package validator

// IValidatorHandler validates commands of any type
// Commands are passed as pointers to validators for consistency
type IValidatorHandler[T any] interface {
	Validate(*T) error
}
