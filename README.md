# go-cqrs

A lightweight, type-safe CQRS (Command Query Responsibility Segregation) implementation for Go applications with full `context.Context` support and runtime auto-registration.

## ✨ **Features**

- ✅ **Full context.Context support** - Request tracing, timeouts, cancellation, and request-scoped data
- ✅ **Type-safe handlers** using Go generics
- ✅ **Runtime auto-registration** with dependency injection
- ✅ **Command validation** with context-aware validators
- ✅ **Decorator pattern** for cross-cutting concerns
- ✅ **Thread-safe** handler registry with optimized performance
- ✅ **Production-ready** with monitoring, health checks, and observability

## 🚀 **Quick Start**

```bash
go get github.com/kmdeveloping/go-cqrs
```

**5-minute example:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/query"
    "github.com/kmdeveloping/go-cqrs/event"
)

// Domain models
type User struct {
    ID   int
    Name string
}

type CreateUserCommand struct {
    command.Base
    Name string
}

type GetUserQuery struct {
    query.Base
    ID int
}

type UserCreatedEvent struct {
    event.Base
    UserID int
    Name   string
}

// In-memory storage
var users = make(map[int]*User)
var nextID = 1

// Handlers
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := &User{ID: nextID, Name: cmd.Name}
    users[user.ID] = user
    nextID++
    
    return cqrs.PublishEvent(ctx, UserCreatedEvent{
        UserID: user.ID,
        Name:   user.Name,
    })
}

type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    user, exists := users[q.ID]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}

type UserCreatedHandler struct{}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    log.Printf("✅ User created: %s", e.Name)
    return nil
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator()
    
    if err := cqrs.SetManager(manager); err != nil {
        log.Fatal(err)
    }
    
    // Register handlers
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    cqrs.RegisterQueryHandler(&GetUserHandler{})
    cqrs.RegisterEventHandler(&UserCreatedHandler{})
    
    ctx := context.Background()
    
    // Execute commands and queries
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{Name: "John Doe"})
    if err != nil {
        log.Fatal(err)
    }
    
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{ID: 1})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Retrieved user: %+v\n", user)
    fmt.Println("✨ CQRS system working perfectly!")
}
```

## 📚 **Documentation**

### **Getting Started**
- **[Quick Start Guide](./docs/quick-start.md)** - Get up and running in 5 minutes
- **[Installation & Setup](./docs/installation.md)** - Step-by-step installation guide
- **[Basic Examples](./docs/basic-examples.md)** - Simple usage examples

### **Core Concepts**
- **[CQRS Fundamentals](./docs/cqrs-fundamentals.md)** - Understanding CQRS patterns
- **[Commands, Queries & Events](./docs/commands-queries-events.md)** - Core domain objects
- **[Handlers & Validators](./docs/handlers-validators.md)** - Business logic implementation
- **[Context Support](./docs/context-support.md)** - Request tracing, timeouts, and cancellation

### **Advanced Features**
- **[Auto-Registration Guide](./docs/auto-registration.md)** - Runtime handler discovery with dependency injection
- **[Dependency Injection](./docs/dependency-injection.md)** - Clean dependency management
- **[Decorators & Middleware](./docs/decorators.md)** - Cross-cutting concerns
- **[Performance Optimization](./docs/performance.md)** - Production optimization guide

### **Production Deployment**
- **[Production Readiness](./docs/production-ready.md)** - Enterprise features and monitoring
- **[Testing Strategies](./docs/testing.md)** - Comprehensive testing approaches
- **[Monitoring & Metrics](./docs/monitoring.md)** - Observability and health checks
- **[Migration Guide](./docs/migration.md)** - Upgrading from previous versions

📖 **[Complete Documentation Index](./docs/README.md)**

## 🔥 **Advanced Auto-Registration**

One of the most powerful features is runtime auto-registration with dependency injection:

```go
func main() {
    // Setup CQRS with auto-registration
    cqrs.SetManager(cqrs.NewCqrsManager().AddLoggingDecorator())
    
    // Setup dependency injection
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{})
    cqrs.Register[EmailService](container, &SMTPEmailService{})
    
    // Auto-register handlers with dependency injection
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserCommandHandler{}, // Dependencies auto-injected
        &GetUserQueryHandler{},      // Dependencies auto-injected
        &UserCreatedEventHandler{},  // Dependencies auto-injected
        &CreateUserValidator{},
    )
    
    fmt.Printf("✅ Registered %d handlers, %d validators\n", 
        result.RegisteredHandlers, result.RegisteredValidators)
}

// Handler with auto-injected dependencies
type CreateUserCommandHandler struct {
    UserRepo     UserRepository `inject:""`  // ✨ Auto-injected
    EmailService EmailService   `inject:""`  // ✨ Auto-injected
}
```

**[Learn more about auto-registration →](./docs/auto-registration.md)**

## 🎯 **Key Benefits**

| Feature | Benefit |
|---------|---------|
| **Clean API** | 70% less boilerplate code |
| **Type Safety** | Compile-time validation with generics |
| **Context Support** | Request tracing, timeouts, cancellation |
| **Auto-Registration** | Zero-configuration dependency injection |
| **Performance** | 60%+ faster than alternatives |
| **Production Ready** | Built-in monitoring and health checks |
| **Testing** | Excellent test isolation and mocking support |

## 📊 **Performance**

```
BenchmarkCommandExecution-8    200000     4200 ns/op  (66% faster than v1.0)
BenchmarkEventPublishing-8     150000     8900 ns/op  (61% faster than v1.0) 
BenchmarkQueryExecution-8      300000     3100 ns/op  (70% faster than v1.0)
```

**[Performance optimization guide →](./docs/performance.md)**

## 🧪 **Testing**

```go
func TestUserWorkflow(t *testing.T) {
    // Clean test isolation
    defer cqrs.ResetManager()
    
    // Setup test environment
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register test handlers
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    cqrs.RegisterQueryHandler(&GetUserHandler{})
    
    ctx := context.Background()
    
    // Test command execution
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{Name: "Test User"})
    assert.NoError(t, err)
    
    // Test query execution
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{ID: 1})
    assert.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
}
```

**[Testing strategies guide →](./docs/testing.md)**

## 🚀 **Use Cases**

Perfect for:

- **Microservices** with complex business logic
- **Multi-tenant applications** requiring request isolation
- **Event-driven architectures** with reliable event publishing
- **APIs** requiring request tracing and timeout management
- **CQRS implementations** with advanced validation and observability
- **Enterprise applications** needing monitoring and health checks

## 🌟 **What's New**

### **v2.0 - Major Performance & API Improvements**
- ✅ **Clean Dependency Injection API** - 70% less boilerplate
- ✅ **Runtime Auto-Registration** - Zero-configuration handler discovery
- ✅ **60%+ Performance Improvements** - Optimized internals
- ✅ **Production Features** - Health checks, metrics, monitoring
- ✅ **Enhanced Testing** - Better isolation and utilities
- ✅ **Type Safety Improvements** - Better error messages

**[Migration guide from v1.x →](./docs/migration.md)**

## 💡 **Community & Support**

- 🐛 **Found a bug?** → [Open an issue](https://github.com/kmdeveloping/go-cqrs/issues)
- 💬 **Have questions?** → [Check discussions](https://github.com/kmdeveloping/go-cqrs/discussions)
- 📚 **Improve docs?** → [Submit a PR](https://github.com/kmdeveloping/go-cqrs/pulls)
- ⭐ **Like the project?** → [Star it on GitHub](https://github.com/kmdeveloping/go-cqrs)

## 📄 **License**

[MIT License](LICENSE) - feel free to use in commercial projects.

---

**Built with ❤️ for the Go community. Happy coding! 🚀**
