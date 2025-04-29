package validator

type IValidator interface{}

type IValidatorHandler[T IValidator] interface {
	Validate(T) error
}

type BaseValidatorHandler struct {
	Valid bool
}

var _ IValidator = (*BaseValidatorHandler)(nil)
