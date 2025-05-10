package handlers

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/example/commands"
)

type DoSomethingCommandValidator struct{}

func (v *DoSomethingCommandValidator) Validate(cmd commands.DoSomethingCommand) error {
	if len(cmd.Something) < 6 {
		return fmt.Errorf("parameter [Something] must have at least 6 characters")
	}

	return nil
}
