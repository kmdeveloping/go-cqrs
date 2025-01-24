package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/manager"
	"github.com/kmdeveloping/go-cqrs/example/contracts/commands"
	"github.com/kmdeveloping/go-cqrs/example/contracts/events"
	"github.com/kmdeveloping/go-cqrs/example/contracts/queries"
)

var cqrsManager manager.ICqrsManager

func init() {
	cqrsManager = manager.NewCqrsManager()
}

func main() {

	cmd := &commands.DoSomethingCommand{
		CustomerNumber: "987098798re8",
	}

	if e := cqrsManager.Execute(cmd); e != nil {
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

	evnt := &events.SomeEventOne{
		Name: "Superman",
		EventBase: &event.EventBase{
			CorrelationUid: uuid.New(),
			ExecutionTime:  time.Now(),
		},
	}

	if err := cqrsManager.Publish(evnt); err != nil {
		fmt.Println(err.Error())
	}
}
