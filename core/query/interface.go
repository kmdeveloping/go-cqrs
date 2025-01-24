package query

type IQuery interface {
}
type QueryBase[TResult any] struct {
	Result TResult
}

var _ IQuery = (*QueryBase[any])(nil)
