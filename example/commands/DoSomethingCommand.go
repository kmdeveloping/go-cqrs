package commands

import "github.com/kmdeveloping/go-cqrs/command"

type DoSomethingCommand struct {
	Something string
	Result    string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
