package command

type ICommand interface {
	GetResult() any
	SetResult(any)
}

type ICommandHandler[T ICommand] interface {
	Handle(*T) error
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
