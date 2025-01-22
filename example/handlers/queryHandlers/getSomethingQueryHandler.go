package queryHandlers

import (
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/example/contracts/queries"
)

// GetSomethingQueryHandler query specific handler extending base handler with concrete types
type GetSomethingQueryHandler struct {
	query.BaseQueryHandler[queries.GetSomethingQuery, []string] `di.inject:"queryHandler"`
}

// non-public handler execute function
func (q *GetSomethingQueryHandler) execute(query *queries.GetSomethingQuery) (*[]string, error) {
	list := []string{query.CustomerNumber, query.CustomerNumber, query.CustomerNumber}
	return &list, nil
}

// NewGetSomethingQuery public getter to access Execute function
func NewGetSomethingQuery() *GetSomethingQueryHandler {
	handler := &GetSomethingQueryHandler{}
	handler.Execute = handler.execute
	return handler
}
