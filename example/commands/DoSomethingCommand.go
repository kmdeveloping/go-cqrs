package commands

import "github.com/kmdeveloping/go-cqrs/command"

type DoSomethingCommand struct {
	command.Base
	Something string
	result    any // private field to store the result
}

// To satisfy the command.ICommand interface
func (c *DoSomethingCommand) GetResult() any {
	return c.result
}

func (c *DoSomethingCommand) SetResult(result any) {
	c.result = result
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
