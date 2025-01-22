package query

type IQuery interface{}

type BaseQueryHandler[TQuery IQuery, TResult any] struct {
	Execute func(*TQuery) (*TResult, error)
}
