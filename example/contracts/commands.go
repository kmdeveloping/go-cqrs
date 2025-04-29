package contracts

import "github.com/kmdeveloping/go-cqrs/core/command"

type DoSomethingCommand struct {
	*command.Base
	Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
