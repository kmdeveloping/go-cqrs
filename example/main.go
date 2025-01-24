package main

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/manager"
	"github.com/kmdeveloping/go-cqrs/example/contracts/commands"
	"github.com/kmdeveloping/go-cqrs/example/handlers/commandHandlers"
)

func main() {

	cqrs := manager.NewCqrsManager()
	if e := cqrs.ExecuteTask(&commandHandlers.DoSomethingCommandHandler{}, &commands.DoSomethingCommand{CustomerNumber: "4563490848"}); e != nil {
		fmt.Println(e.Error())
	}
}
