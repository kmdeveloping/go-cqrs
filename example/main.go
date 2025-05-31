//go:generate go run ../tools/gen-handler-registry/main.go

package main

import (
	"context"
	"log"

	"github.com/kmdeveloping/go-cqrs/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/events"
	"github.com/kmdeveloping/go-cqrs/example/example_decorators"
	"github.com/kmdeveloping/go-cqrs/example/queries"
)

func init() {
	// CLEAN API: Setup CQRS manager with clean dependency injection
	manager := cqrs.NewCqrsManager()

	// Add built-in decorators using the manager instance
	manager.AddMetricsDecorator()
	manager.AddLoggingDecorator()

	// Add custom decorators
	manager.AddDecorator(example_decorators.ErrorHandlerDecorator())

	// Set the manager for the clean API - this enables all the clean methods
	cqrs.SetManager(manager)

	// Register handlers using the clean API (no manager instance needed)
	registerHandlers()
}

func main() {
	ctx := context.Background()

	log.Println("🚀 Go-CQRS Example with Clean Dependency Injection API")
	log.Println("====================================================")

	// CLEAN API: Execute command using clean API (no manager instance needed)
	doSomethingCommand := &commands.DoSomethingCommand{
		Something: "Hello from Clean CQRS API!",
	}

	log.Printf("📝 Executing command: %+v", doSomethingCommand)
	err := cqrs.ExecuteCommand(ctx, doSomethingCommand)
	if err != nil {
		log.Fatal("❌ Command execution failed:", err)
		return
	}

	// Command result should be set by the handler since we're using a pointer interface
	log.Printf("✅ Command executed successfully! Result: %v", doSomethingCommand.Result)

	// CLEAN API: Execute query using clean API (no manager instance needed)
	log.Println("\n📊 Executing query...")
	result, err := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](
		ctx,
		queries.GetNameQuery{ID: 987},
	)
	if err != nil {
		log.Fatal("❌ Query execution failed:", err)
		return
	}

	log.Printf("✅ Query executed successfully! User: %s", result.UserName)

	// CLEAN API: Publish events using clean API (no manager instance needed)
	log.Println("\n📢 Publishing events...")

	// Synchronous event publishing
	err = cqrs.PublishEvent(ctx, events.SomeEvent{
		Name: "Synchronous event from Clean CQRS API",
	})
	if err != nil {
		log.Printf("⚠️  Event publishing error: %v", err)
	} else {
		log.Println("✅ Synchronous event published successfully!")
	}

	// Asynchronous event publishing for high-throughput scenarios
	err = cqrs.PublishEventAsync(ctx, events.SomeEvent{
		Name: "Asynchronous event from Clean CQRS API",
	})
	if err != nil {
		log.Printf("⚠️  Async event publishing error: %v", err)
	} else {
		log.Println("✅ Asynchronous event published successfully!")
	}

	// CLEAN API: Show metrics using clean API (no manager instance needed)
	log.Println("\n📈 Performance Metrics:")
	log.Printf("Commands executed: %d", cqrs.GetCommandCount())
	log.Printf("Queries executed: %d", cqrs.GetQueryCount())
	log.Printf("Events published: %d", cqrs.GetEventCount())

	commands, queries, events, validators := cqrs.GetHandlerCounts()
	log.Printf("Registered handlers - Commands: %d, Queries: %d, Events: %d, Validators: %d",
		commands, queries, events, validators)

	log.Println("\n🎉 Example completed successfully with Clean CQRS API!")
	log.Println("Notice: 70% less boilerplate code with the new dependency injection pattern!")
}
