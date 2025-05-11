package command

type ICommand any

type ICommandHandler[T ICommand] interface {
	Handle(T) error
}
type Base struct{}

var _ ICommand = (*Base)(nil)
