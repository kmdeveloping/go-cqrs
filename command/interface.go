package command

type ICommand interface{}

type CommandHandler[TCommand ICommand] interface {
	Execute(TCommand) error
}
