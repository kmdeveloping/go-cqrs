package query

type IQuery interface{}

type QueryHandler[TQuery IQuery, TResult any] interface {
	Execute(TQuery) (TResult, error)
}
