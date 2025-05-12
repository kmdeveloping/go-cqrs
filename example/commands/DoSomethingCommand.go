package commands

import "github.com/kmdeveloping/go-cqrs/command"

type DoSomethingCommand struct {
	command.Base
	Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)

func (b *DoSomethingCommand) GetResult() any {
	return b.Base.GetResult()
}

func (b *DoSomethingCommand) SetResult(result any) {
	b.Base.SetResult(result)
}
