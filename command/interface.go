package command

type ICommand any

// ICommandHandler interface handles commands of type T that implements ICommand
// The T can be either a value or a pointer type that implements ICommand
type ICommandHandler[T any] interface {
	Handle(T) error
}

type Base struct{}

var _ ICommand = (*Base)(nil)
