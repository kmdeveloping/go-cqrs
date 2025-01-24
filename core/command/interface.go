package command

type ICommand interface{}

type CommandBase struct{}

var _ ICommand = (*CommandBase)(nil)
