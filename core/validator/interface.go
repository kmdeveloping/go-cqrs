package validator

import "github.com/kmdeveloping/go-cqrs/core/command"

type IValidatorHandler[T command.ICommand] interface {
	Validate(T) error
}
