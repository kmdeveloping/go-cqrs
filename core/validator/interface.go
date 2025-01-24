package validator

type IValidator interface{}

type BaseValidatorHandler struct {
	Valid bool
}

var _ IValidator = (*BaseValidatorHandler)(nil)
