package validator

// IValidatorHandler validates commands of any type
type IValidatorHandler[T any] interface {
	Validate(T) error
}
