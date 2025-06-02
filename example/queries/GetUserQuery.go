package queries

import "github.com/kmdeveloping/go-cqrs/query"

// GetUserQuery demonstrates a query for the dependency injection example
type GetUserQuery struct {
	query.Base
	ID int
}

// GetUserQueryResponse represents the response for GetUserQuery
type GetUserQueryResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var _ query.IQuery = (*GetUserQuery)(nil)
