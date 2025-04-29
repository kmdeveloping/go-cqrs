package handler

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type DoSomethingCommandValidator struct{}

func (v *DoSomethingCommandValidator) Validate(cmd contracts.DoSomethingCommand) error {
	if len(cmd.Something) < 6 {
		return fmt.Errorf("parameter [Something] must have at least 6 characters")
	}

	return nil
}
