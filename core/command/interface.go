package command

type ICommand interface{}
type ICommandWithResult interface {
	ICommand
}

type CommandBase struct{}
type CommandWithResultBase[TResult any] struct {
	Result TResult
}
