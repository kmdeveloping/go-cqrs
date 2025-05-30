package command

import "context"

type ICommand any

// ICommandHandler interface handles commands of type T that implements ICommand
// Commands are always passed as pointers to handlers for consistency
type ICommandHandler[T ICommand] interface {
	Handle(ctx context.Context, cmd *T) error
}

type Base struct{}
type BaseWithResult struct {
	Result any
}

var _ ICommand = (*Base)(nil)
var _ ICommand = (*BaseWithResult)(nil)
