package commands

import "github.com/kmdeveloping/go-cqrs/command"

type DoSomethingCommand struct {
	command.BaseWithResult
	Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
