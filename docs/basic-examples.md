# Basic Examples

Simple, practical examples to get you started with go-cqrs.

## 📚 Table of Contents

- [Simple Command Example](#simple-command-example)
- [Query with Response Example](#query-with-response-example)
- [Event Publishing Example](#event-publishing-example)
- [Complete CRUD Example](#complete-crud-example)
- [Validation Example](#validation-example)
- [Context Usage Example](#context-usage-example)

## 🎯 Simple Command Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
)

// Command definition
type CreateUserCommand struct {
    command.Base
    Name  string
    Email string
}

// Command handler
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Simulate user creation
    fmt.Printf("Creating user: %s (%s)\n", cmd.Name, cmd.Email)
    
    // In real app, save to database here
    // userID, err := h.userService.Create(ctx, cmd.Name, cmd.Email)
    
    return nil
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register handler
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    
    // Execute command
    ctx := context.Background()
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ User created successfully!")
}
```

## 🔍 Query with Response Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/query"
)

// Domain model
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Query definition
type GetUserQuery struct {
    query.Base
    UserID int
}

// Query handler
type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    // Simulate database lookup
    // In real app: return h.userRepo.GetByID(ctx, q.UserID)
    
    user := &User{
        ID:    q.UserID,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    return user, nil
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register handler
    cqrs.RegisterQueryHandler(&GetUserHandler{})
    
    // Execute query
    ctx := context.Background()
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{
        UserID: 123,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("✅ Found user: %+v\n", user)
}
```

## 📢 Event Publishing Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/google/uuid"
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/event"
)

// Event definition
type UserCreatedEvent struct {
    event.Base
    UserID int    `json:"user_id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
}

// Event handler
type UserCreatedHandler struct{}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    fmt.Printf("🎉 Event received: User %s created with ID %d\n", e.Name, e.UserID)
    
    // Handle the event (send email, update analytics, etc.)
    return nil
}

// Command that publishes event
type CreateUserCommand struct {
    command.Base
    Name  string
    Email string
}

type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Create user logic here...
    userID := 123 // Simulate created user ID
    
    // Publish event
    userCreatedEvent := UserCreatedEvent{
        Base: event.Base{
            ExecutionTime:  time.Now(),
            CorrelationUid: uuid.New(),
            MetaData:       "user-creation",
        },
        UserID: userID,
        Name:   cmd.Name,
        Email:  cmd.Email,
    }
    
    return cqrs.PublishEvent(ctx, userCreatedEvent)
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register handlers
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    cqrs.RegisterEventHandler(&UserCreatedHandler{})
    
    // Execute command (which will publish event)
    ctx := context.Background()
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "Alice Smith",
        Email: "alice@example.com",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ Command executed and event published!")
}
```

## 📋 Complete CRUD Example

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "sync"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/query"
)

// Domain model
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// In-memory storage (use database in real app)
var (
    users   = make(map[int]*User)
    nextID  = 1
    usersMu sync.RWMutex
)

// Commands
type CreateUserCommand struct {
    command.Base
    Name  string
    Email string
}

type UpdateUserCommand struct {
    command.Base
    ID    int
    Name  string
    Email string
}

type DeleteUserCommand struct {
    command.Base
    ID int
}

// Queries
type GetUserQuery struct {
    query.Base
    ID int
}

type ListUsersQuery struct {
    query.Base
    Limit  int
    Offset int
}

// Command Handlers
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    usersMu.Lock()
    defer usersMu.Unlock()
    
    user := &User{
        ID:    nextID,
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    users[nextID] = user
    nextID++
    
    fmt.Printf("✅ Created user: %+v\n", user)
    return nil
}

type UpdateUserHandler struct{}

func (h *UpdateUserHandler) Handle(ctx context.Context, cmd *UpdateUserCommand) error {
    usersMu.Lock()
    defer usersMu.Unlock()
    
    user, exists := users[cmd.ID]
    if !exists {
        return errors.New("user not found")
    }
    
    user.Name = cmd.Name
    user.Email = cmd.Email
    
    fmt.Printf("✅ Updated user: %+v\n", user)
    return nil
}

type DeleteUserHandler struct{}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd *DeleteUserCommand) error {
    usersMu.Lock()
    defer usersMu.Unlock()
    
    if _, exists := users[cmd.ID]; !exists {
        return errors.New("user not found")
    }
    
    delete(users, cmd.ID)
    fmt.Printf("✅ Deleted user with ID: %d\n", cmd.ID)
    return nil
}

// Query Handlers
type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    usersMu.RLock()
    defer usersMu.RUnlock()
    
    user, exists := users[q.ID]
    if !exists {
        return nil, errors.New("user not found")
    }
    
    return user, nil
}

type ListUsersHandler struct{}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) ([]*User, error) {
    usersMu.RLock()
    defer usersMu.RUnlock()
    
    var result []*User
    count := 0
    
    for _, user := range users {
        if count >= q.Offset && len(result) < q.Limit {
            result = append(result, user)
        }
        count++
    }
    
    return result, nil
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator()
    cqrs.SetManager(manager)
    
    // Register handlers
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    cqrs.RegisterCommandHandler(&UpdateUserHandler{})
    cqrs.RegisterCommandHandler(&DeleteUserHandler{})
    cqrs.RegisterQueryHandler(&GetUserHandler{})
    cqrs.RegisterQueryHandler(&ListUsersHandler{})
    
    ctx := context.Background()
    
    // CREATE
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    err = cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "Jane Smith",
        Email: "jane@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // READ
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{ID: 1})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("📖 Retrieved user: %+v\n", user)
    
    // UPDATE
    err = cqrs.ExecuteCommand(ctx, &UpdateUserCommand{
        ID:    1,
        Name:  "John Updated",
        Email: "john.updated@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // LIST
    usersList, err := cqrs.ExecuteQuery[ListUsersQuery, []*User](ctx, ListUsersQuery{
        Limit:  10,
        Offset: 0,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("📋 All users: %+v\n", usersList)
    
    // DELETE
    err = cqrs.ExecuteCommand(ctx, &DeleteUserCommand{ID: 2})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("🎉 CRUD operations completed successfully!")
}
```

## ✅ Validation Example

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "strings"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/validator"
)

type CreateUserCommand struct {
    command.Base
    Name  string
    Email string
}

// Validator
type CreateUserValidator struct{}

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    if strings.TrimSpace(cmd.Name) == "" {
        return errors.New("name is required")
    }
    
    if len(cmd.Name) < 2 {
        return errors.New("name must be at least 2 characters")
    }
    
    if !strings.Contains(cmd.Email, "@") {
        return errors.New("invalid email format")
    }
    
    return nil
}

// Handler (only called if validation passes)
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    fmt.Printf("✅ Creating validated user: %s (%s)\n", cmd.Name, cmd.Email)
    return nil
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register validator and handler
    cqrs.RegisterCommandValidator(&CreateUserValidator{})
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    
    ctx := context.Background()
    
    // Valid command
    fmt.Println("Testing valid command:")
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
    }
    
    // Invalid command (empty name)
    fmt.Println("\nTesting invalid command (empty name):")
    err = cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "",
        Email: "john@example.com",
    })
    if err != nil {
        fmt.Printf("❌ Validation failed: %v\n", err)
    }
    
    // Invalid command (bad email)
    fmt.Println("\nTesting invalid command (bad email):")
    err = cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "John Doe",
        Email: "invalid-email",
    })
    if err != nil {
        fmt.Printf("❌ Validation failed: %v\n", err)
    }
}
```

## 🕐 Context Usage Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
)

type ProcessDataCommand struct {
    command.Base
    Data string
}

type ProcessDataHandler struct{}

func (h *ProcessDataHandler) Handle(ctx context.Context, cmd *ProcessDataCommand) error {
    // Check if context has timeout
    deadline, hasDeadline := ctx.Deadline()
    if hasDeadline {
        fmt.Printf("⏰ Processing with deadline: %v\n", deadline)
    }
    
    // Simulate long-running operation
    select {
    case <-time.After(2 * time.Second):
        fmt.Printf("✅ Processed data: %s\n", cmd.Data)
        return nil
    case <-ctx.Done():
        fmt.Printf("❌ Processing cancelled: %v\n", ctx.Err())
        return ctx.Err()
    }
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register handler
    cqrs.RegisterCommandHandler(&ProcessDataHandler{})
    
    // Example 1: Normal execution
    fmt.Println("=== Normal Execution ===")
    ctx := context.Background()
    err := cqrs.ExecuteCommand(ctx, &ProcessDataCommand{Data: "sample data"})
    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Example 2: With timeout (will succeed)
    fmt.Println("\n=== With Timeout (3s - should succeed) ===")
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    err = cqrs.ExecuteCommand(ctx, &ProcessDataCommand{Data: "timeout test data"})
    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Example 3: With short timeout (will fail)
    fmt.Println("\n=== With Short Timeout (1s - should fail) ===")
    ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    err = cqrs.ExecuteCommand(ctx, &ProcessDataCommand{Data: "short timeout data"})
    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Example 4: With cancellation
    fmt.Println("\n=== With Manual Cancellation ===")
    ctx, cancel = context.WithCancel(context.Background())
    
    // Cancel after 1 second
    go func() {
        time.Sleep(1 * time.Second)
        fmt.Println("🛑 Cancelling context...")
        cancel()
    }()
    
    err = cqrs.ExecuteCommand(ctx, &ProcessDataCommand{Data: "cancellation test"})
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
```

## 🚀 Next Steps

1. **Learn Core Concepts:** [Commands, Queries & Events](./commands-queries-events.md)
2. **Add Validation:** [Handlers & Validators](./handlers-validators.md)
3. **Use Auto-Registration:** [Auto-Registration Guide](./auto-registration.md)
4. **Production Setup:** [Production Readiness](./production-ready.md)

---

**Ready to dive deeper? Explore [CQRS Fundamentals](./cqrs-fundamentals.md)! 📚** 