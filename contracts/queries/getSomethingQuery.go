package queries

import "github.com/kmdeveloping/go-cqrs/query"

type GetSomethingQuery struct {
	CustomerNumber string
}

type GetSomethingQueryResponse struct {
	Result []string
}

func (q *GetSomethingQuery) Execute(query *GetSomethingQuery) (*GetSomethingQueryResponse, error) {
	return &GetSomethingQueryResponse{
		Result: []string{query.CustomerNumber, query.CustomerNumber, query.CustomerNumber},
	}, nil
}

func NewGetSomethingQueryHandler() query.QueryHandler[*GetSomethingQuery, *GetSomethingQueryResponse] {
	return &GetSomethingQuery{}
}
