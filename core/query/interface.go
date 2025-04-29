package query

type IQuery interface{}

type IQueryHandler[T IQuery, R any] interface {
	Handle(T) (R, error)
}
type Base struct{}

var _ IQuery = (*Base)(nil)
