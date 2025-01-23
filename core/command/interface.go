package command

type ICommand interface{}

type BaseCommandHandler[TCommand ICommand] struct {
	Execute func(*TCommand) error
}
