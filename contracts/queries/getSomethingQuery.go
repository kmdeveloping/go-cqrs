package queries

import "github.com/kmdeveloping/go-cqrs/query"

// Query contract
type GetSomethingQuery struct {
	CustomerNumber string
}

// query specific handler extending base handler with concrete types
type GetSomethingQueryHandler struct {
	query.BaseQueryHandler[GetSomethingQuery, []string]
}

// non public handler execute function
func (q *GetSomethingQueryHandler) execute(query *GetSomethingQuery) (*[]string, error) {
	list := []string{query.CustomerNumber, query.CustomerNumber, query.CustomerNumber}
	return &list, nil
}

// public getter to access Execute function
func NewGetSomethingQuery() *GetSomethingQueryHandler {
	handler := &GetSomethingQueryHandler{}
	handler.Execute = handler.execute
	return handler
}
