package handler

import (
	"errors"
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type GetNewUserQueryHandler struct {
	*handlers.BaseHandler
}

var _ handlers.IQueryHandler[any] = (*GetNewUserQueryHandler)(nil)

func (q *GetNewUserQueryHandler) Get(qry query.IQuery) (any, error) {
	qr, ok := qry.(contracts.GetNewUserQuery)
	if !ok {
		return "", errors.New("invalid query type")
	}

	result := fmt.Sprintf("%s%s@%s", string(qr.FirstName[0]), qr.LastName, qr.Project)
	return result, nil
}
