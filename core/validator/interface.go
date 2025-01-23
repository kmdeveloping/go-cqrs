package validator

type IValidator interface{}

type BaseValidatorHandler[TValidator IValidator] struct {
	Validate func(*TValidator) error
}
