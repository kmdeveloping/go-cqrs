package command

type ICommand interface {
	GetResult() any
	SetResult(any)
}

// ICommandHandler interface handles commands of type T that implements ICommand
// The T can be either a value or a pointer type that implements ICommand
type ICommandHandler[T any] interface {
	Handle(T) error
}

type Base struct {
	Result any
}

func (b *Base) GetResult() any {
	return b.Result
}

func (b *Base) SetResult(result any) {
	b.Result = result
}

var _ ICommand = (*Base)(nil)
