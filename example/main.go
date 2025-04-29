package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
	"log"
)

var CqrsManager *cqrs.Manager

func init() {

	CqrsManager = cqrs.NewCqrsManager()
	//CqrsManager.UseDefaultDecorators()

	cqrs.RegisterCommandHandler(CqrsManager, &handler.DoThatCommandHandler{})
	cqrs.RegisterQueryHandler(CqrsManager, &handler.GetNameQueryHandler{})
}

func main() {
	err := cqrs.ExecuteCommand(CqrsManager, contracts.DoSomethingCommand{Something: "Hello"})
	if err != nil {
		log.Fatal(err)
		return
	}

	result, err := cqrs.ExecuteQuery[contracts.GetNameQuery, contracts.GetNameQueryResponse](CqrsManager, contracts.GetNameQuery{ID: 987})
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println(result.UserName)
}
