package commands

import "github.com/kmdeveloping/go-cqrs/command"

type DoSomethingCommand struct {
	Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
