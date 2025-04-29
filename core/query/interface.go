package query

type IQuery interface{}

type IQueryHandler[T IQuery, R any] interface {
	Get(T) (R, error)
}
type QueryBase struct{}
