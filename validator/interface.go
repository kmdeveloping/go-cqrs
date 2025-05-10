package validator

import "github.com/kmdeveloping/go-cqrs/command"

type IValidatorHandler[T command.ICommand] interface {
	Validate(T) error
}
