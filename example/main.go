package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
	"log"
)

var dispatch *cqrs.Manager

func init() {
	config := &cqrs.Configuration{}
	config.UseLoggingDecorator()
	config.UseMetricsDecorator()
	config.UseErrorHandlerDecorator()

	dispatch = cqrs.NewCqrsManager(config)

	cqrs.RegisterCommandHandler(dispatch, &handler.DoThatCommandHandler{})
}

func main() {
	err := cqrs.ExecuteCommand(dispatch, contracts.DoSomethingCommand{Something: "Hello"})
	if err != nil {
		log.Fatal(err)
		return
	}
}
