package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
	"log"
)

var mgr *cqrs.Manager

func init() {

	mgr = cqrs.NewCqrsManager()
	mgr.UseDefaultDecorators()

	cqrs.RegisterCqrsManager(mgr)
	cqrs.RegisterCommandHandler(mgr, cqrs.NewHandler[handler.DoThatCommandHandler]())
	cqrs.RegisterQueryHandler(mgr, cqrs.NewHandler[handler.GetNameQueryHandler]())
	cqrs.RegisterEventHandler(mgr, &handler.SomeEventHandler{})
	cqrs.RegisterEventHandler(mgr, &handler.SomeOtherEventHandler{})
	cqrs.RegisterValidator(mgr, &handler.DoSomethingCommandValidator{})
}

func main() {
	err := cqrs.ExecuteCommand(mgr, contracts.DoSomethingCommand{Something: "Hello Something it is nice to see you"})
	if err != nil {
		log.Fatal(err)
		return
	}

	result, er := cqrs.ExecuteQuery[contracts.GetNameQuery, contracts.GetNameQueryResponse](mgr, contracts.GetNameQuery{ID: 987})
	if er != nil {
		log.Fatal(er)
		return
	}

	log.Println(result.UserName)
}
