package handler

import (
	"errors"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
)

type GetNameQueryHandler struct{}

func (h *GetNameQueryHandler) Handle(qry contracts.GetNameQuery) (contracts.GetNameQueryResponse, error) {
	if qry.ID >= 37 {
		return contracts.GetNameQueryResponse{
			ID:       qry.ID,
			UserName: "YouHaveReturnedAQuery",
		}, nil
	}

	return contracts.GetNameQueryResponse{}, errors.New("user id not found")
}
