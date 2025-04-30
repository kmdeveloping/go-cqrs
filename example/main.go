package main

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/handler"
	"github.com/kmdeveloping/go-cqrs/example/queries"
)

var mgr *cqrs.Manager

func init() {

	mgr = cqrs.NewCqrsManager()
	//mgr.UseDefaultDecorators()
	mgr.AddLoggingDecorator()
	mgr.AddMetricsDecorator()

	cqrs.RegisterCqrsManager(mgr)
	cqrs.RegisterCommandHandler(mgr, cqrs.NewHandler[handler.DoThatCommandHandler]())
	cqrs.RegisterQueryHandler(mgr, cqrs.NewHandler[handler.GetNameQueryHandler]())
	cqrs.RegisterEventHandler(mgr, &handler.SomeEventHandler{})
	cqrs.RegisterEventHandler(mgr, &handler.SomeOtherEventHandler{})
	cqrs.RegisterValidator(mgr, &handler.DoSomethingCommandValidator{})
}

func main() {
	err := cqrs.ExecuteCommand(mgr, commands.DoSomethingCommand{Something: "Hello Something it is nice to see you"})
	if err != nil {
		log.Fatal(err)
		return
	}

	result, er := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](mgr, queries.GetNameQuery{ID: 987})
	if er != nil {
		log.Fatal(er)
		return
	}

	log.Println(result.UserName)
}
