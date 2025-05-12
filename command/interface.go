package command

type ICommand any

// ICommandHandler interface handles commands of type T that implements ICommand
// Commands are always passed as pointers to handlers for consistency
type ICommandHandler[T ICommand] interface {
	Handle(*T) error
}

type Base struct{}

var _ ICommand = (*Base)(nil)
