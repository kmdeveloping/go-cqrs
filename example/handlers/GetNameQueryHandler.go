package handlers

import (
	"errors"

	"github.com/kmdeveloping/go-cqrs/example/queries"
)

type GetNameQueryHandler struct{}

func (h GetNameQueryHandler) Handle(qry queries.GetNameQuery) (queries.GetNameQueryResponse, error) {
	if qry.ID >= 37 {
		return queries.GetNameQueryResponse{
			ID:       qry.ID,
			UserName: "YouHaveReturnedAQuery",
		}, nil
	}

	return queries.GetNameQueryResponse{}, errors.New("user id not found")
}
