package contracts

import "github.com/kmdeveloping/go-cqrs/core/command"

type DoSomethingCommand struct {
	*command.CommandBase
	Something string
}
