package contracts

import "github.com/kmdeveloping/go-cqrs/core/command"

type DoSomethingCommand struct {
	*command.CommandBase
	Something string
}

type DoSomethingWithResultCommand struct {
	*command.CommandWithResultBase[string]
	Something string
}
