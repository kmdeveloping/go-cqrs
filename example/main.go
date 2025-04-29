package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
	"log"
)

var dispatch *cqrs.CqrsManager

func init() {
	config := &cqrs.CqrsConfiguration{}
	config.UseLoggingDecorator()
	config.UseMetricsDecorator()
	config.UseErrorHandlerDecorator()

	dispatch = cqrs.NewCqrsManager(config)

	cqrs.RegisterHandler(dispatch, &handler.DoThatCommandHandler{})
}

func main() {
	err := cqrs.Execute(dispatch, contracts.DoSomethingCommand{Something: "Hello"})
	if err != nil {
		log.Fatal(err)
		return
	}
}
