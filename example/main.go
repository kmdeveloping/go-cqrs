package main

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/manager"
	"github.com/kmdeveloping/go-cqrs/example/contracts/commands"
	"github.com/kmdeveloping/go-cqrs/example/contracts/queries"
)

var cqrsManager manager.ICqrsManager

func init() {
	cqrsManager = manager.NewCqrsManager()
}

func main() {

	cmd := commands.DoSomethingCommand{
		CustomerNumber: "987098798re8",
	}

	if e := cqrsManager.Execute(&cmd); e != nil {
		fmt.Println(e.Error())
	}

	qry := &queries.GetSomethingQuery{
		CustomerNumber: "ooieurjnavkun8",
	}

	if err := cqrsManager.Execute(qry); err != nil {
		fmt.Println(err.Error())
	}

	for _, r := range qry.Result {
		fmt.Println(r)
	}
}
