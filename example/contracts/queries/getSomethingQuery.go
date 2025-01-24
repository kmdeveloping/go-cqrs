package queries

import "github.com/kmdeveloping/go-cqrs/core/query"

// GetSomethingQuery Query contract
type GetSomethingQuery struct {
	query.QueryBase[[]string]
	CustomerNumber string
}

var _ query.IQuery = (*GetSomethingQuery)(nil)

// non-public handler execute function
func (q *GetSomethingQuery) Handler(query *GetSomethingQuery) error {
	q.Result = []string{query.CustomerNumber, query.CustomerNumber, query.CustomerNumber}
	return nil
}
