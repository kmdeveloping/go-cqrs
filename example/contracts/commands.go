package contracts

import "github.com/kmdeveloping/go-cqrs/core/command"

type DoSomethingCommand struct {
	Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
