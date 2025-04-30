package main

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/handler"
	"github.com/kmdeveloping/go-cqrs/example/queries"
)

func init() {

	m := cqrs.NewCqrsManager()
	m.UseDefaultDecorators()

	cqrs.RegisterCqrsManager()
	cqrs.RegisterCommandHandler(cqrs.NewHandler[handler.DoThatCommandHandler]())
	cqrs.RegisterQueryHandler(cqrs.NewHandler[handler.GetNameQueryHandler]())
	cqrs.RegisterEventHandler(&handler.SomeEventHandler{})
	cqrs.RegisterEventHandler(&handler.SomeOtherEventHandler{})
	cqrs.RegisterValidator(&handler.DoSomethingCommandValidator{})
}

func main() {
	err := cqrs.ExecuteCommand(commands.DoSomethingCommand{Something: "Helloooooo"})
	if err != nil {
		log.Fatal(err)
		return
	}

	result, er := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](queries.GetNameQuery{ID: 987})
	if er != nil {
		log.Fatal(er)
		return
	}

	log.Println(result.UserName)
}
