package query

import "context"

type IQuery any

type IQueryHandler[T IQuery, R any] interface {
	Handle(ctx context.Context, query T) (R, error)
}

type Base struct{}

var _ IQuery = (*Base)(nil)
