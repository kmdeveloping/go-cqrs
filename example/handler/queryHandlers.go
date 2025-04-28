package handler

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type GetNewUserQueryHandler struct {
}

var _ handlers.IQueryHandler[contracts.GetNewUserQuery, string] = (*GetNewUserQueryHandler)(nil)

func (q *GetNewUserQueryHandler) Get(qry contracts.GetNewUserQuery) (string, error) {
	result := fmt.Sprintf("%s%s@%s", string(qry.FirstName[0]), qry.LastName, qry.Project)
	return result, nil
}
