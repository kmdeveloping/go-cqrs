package contracts

import "github.com/kmdeveloping/go-cqrs/core/query"

type GetNewUserQuery struct {
	*query.QueryBase
	FirstName string
	LastName  string
	Project   string
}

var _ query.IQuery = (*GetNewUserQuery)(nil)
