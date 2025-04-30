package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/handler"
)

func BootstrapCqrs() {
	m := cqrs.NewCqrsManager()
	m.UseDefaultDecorators()
	registerHandlers()
}

func registerHandlers() {
	cqrs.RegisterCommandHandler(&handler.DoThatCommandHandler{})
	cqrs.RegisterQueryHandler(&handler.GetNameQueryHandler{})
	cqrs.RegisterEventHandler(&handler.SomeEventHandler{})
	cqrs.RegisterEventHandler(&handler.SomeOtherEventHandler{})
	cqrs.RegisterValidator(&handler.DoSomethingCommandValidator{})
}
