# Quick Start Guide

Get up and running with go-cqrs in 5 minutes! This guide will walk you through creating a simple user management system using CQRS patterns.

## 📦 **Installation**

```bash
go get github.com/kmdeveloping/go-cqrs
```

## 🚀 **5-Minute Setup**

### **Step 1: Define Your Domain Models**

Create your commands, queries, and events:

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

// Domain model
type User struct {
    ID   int
    Name string
    Email string
}

// Command
type CreateUserCommand struct {
    command.Base
    Name  string
    Email string
}

// Query  
type GetUserQuery struct {
    query.Base
    ID int
}

// Event
type UserCreatedEvent struct {
    event.Base
    UserID int
    Name   string
    Email  string
}
```

### **Step 2: Create Handlers**

Implement your business logic:

```go
// In-memory storage for this example
var users = make(map[int]*User)
var nextID = 1

// Command Handler
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := &User{
        ID:    nextID,
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    users[user.ID] = user
    nextID++
    
    // Publish event
    return cqrs.PublishEvent(ctx, UserCreatedEvent{
        UserID: user.ID,
        Name:   user.Name,
        Email:  user.Email,
    })
}

// Query Handler
type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    user, exists := users[q.ID]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}

// Event Handler
type UserCreatedHandler struct{}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    log.Printf("✅ User created: %s (%s)", e.Name, e.Email)
    return nil
}
```

### **Step 3: Setup and Register**

Configure the CQRS manager and register your handlers:

```go
func main() {
    // Setup CQRS manager
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator()  // Optional: adds logging
    
    if err := cqrs.SetManager(manager); err != nil {
        log.Fatal(err)
    }
    
    // Register handlers
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    cqrs.RegisterQueryHandler(&GetUserHandler{})
    cqrs.RegisterEventHandler(&UserCreatedHandler{})
    
    // Test the system
    ctx := context.Background()
    
    // Create a user
    fmt.Println("Creating user...")
    err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Query the user
    fmt.Println("Retrieving user...")
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{ID: 1})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Retrieved user: %+v\n", user)
    
    fmt.Println("✨ Success! Your CQRS system is working!")
}
```

## 🎯 **Complete Working Example**

Here's the full `main.go` file you can copy and run:

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

// Domain model
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Commands
type CreateUserCommand struct {
    command.Base
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Queries
type GetUserQuery struct {
    query.Base
    ID int `json:"id"`
}

type ListUsersQuery struct {
    query.Base
}

// Events
type UserCreatedEvent struct {
    event.Base
    UserID int    `json:"user_id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
}

// Storage
var users = make(map[int]*User)
var nextID = 1

// Handlers
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := &User{
        ID:    nextID,
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    users[user.ID] = user
    nextID++
    
    return cqrs.PublishEvent(ctx, UserCreatedEvent{
        UserID: user.ID,
        Name:   user.Name,
        Email:  user.Email,
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

type ListUsersHandler struct{}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) ([]*User, error) {
    var result []*User
    for _, user := range users {
        result = append(result, user)
    }
    return result, nil
}

type UserCreatedHandler struct{}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    log.Printf("🎉 User created: %s (%s) with ID %d", e.Name, e.Email, e.UserID)
    return nil
}

func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator()
    
    if err := cqrs.SetManager(manager); err != nil {
        log.Fatal("Failed to set manager:", err)
    }
    
    // Register handlers
    cqrs.RegisterCommandHandler(&CreateUserHandler{})
    cqrs.RegisterQueryHandler(&GetUserHandler{})
    cqrs.RegisterQueryHandler(&ListUsersHandler{})
    cqrs.RegisterEventHandler(&UserCreatedHandler{})
    
    ctx := context.Background()
    
    // Create users
    fmt.Println("📝 Creating users...")
    
    users := []CreateUserCommand{
        {Name: "Alice Smith", Email: "alice@example.com"},
        {Name: "Bob Johnson", Email: "bob@example.com"},
        {Name: "Charlie Brown", Email: "charlie@example.com"},
    }
    
    for _, cmd := range users {
        err := cqrs.ExecuteCommand(ctx, &cmd)
        if err != nil {
            log.Printf("❌ Failed to create user %s: %v", cmd.Name, err)
        }
    }
    
    // Query individual user
    fmt.Println("\n🔍 Retrieving individual user...")
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{ID: 1})
    if err != nil {
        log.Printf("❌ Failed to get user: %v", err)
    } else {
        fmt.Printf("✅ Retrieved user: %+v\n", user)
    }
    
    // List all users
    fmt.Println("\n📋 Listing all users...")
    allUsers, err := cqrs.ExecuteQuery[ListUsersQuery, []*User](ctx, ListUsersQuery{})
    if err != nil {
        log.Printf("❌ Failed to list users: %v", err)
    } else {
        fmt.Printf("✅ Found %d users:\n", len(allUsers))
        for _, u := range allUsers {
            fmt.Printf("  - %s (%s)\n", u.Name, u.Email)
        }
    }
    
    // Show metrics
    fmt.Println("\n📊 System metrics:")
    commands := cqrs.GetCommandCount()
    queries := cqrs.GetQueryCount()
    events := cqrs.GetEventCount()
    fmt.Printf("  Commands executed: %d\n", commands)
    fmt.Printf("  Queries executed: %d\n", queries)
    fmt.Printf("  Events published: %d\n", events)
    
    fmt.Println("\n✨ CQRS system is working perfectly!")
}
```

## 🏃‍♂️ **Run It**

```bash
# Create a new Go module
go mod init my-cqrs-app

# Add the dependency
go get github.com/kmdeveloping/go-cqrs

# Copy the code above into main.go

# Run it!
go run main.go
```

**Expected Output:**
```
📝 Creating users...
🎉 User created: Alice Smith (alice@example.com) with ID 1
🎉 User created: Bob Johnson (bob@example.com) with ID 2
🎉 User created: Charlie Brown (charlie@example.com) with ID 3

🔍 Retrieving individual user...
✅ Retrieved user: &{ID:1 Name:Alice Smith Email:alice@example.com}

📋 Listing all users...
✅ Found 3 users:
  - Alice Smith (alice@example.com)
  - Bob Johnson (bob@example.com)
  - Charlie Brown (charlie@example.com)

📊 System metrics:
  Commands executed: 3
  Queries executed: 2
  Events published: 3

✨ CQRS system is working perfectly!
```

## 🎯 **What You've Accomplished**

In just 5 minutes, you've:

✅ **Set up a complete CQRS system**  
✅ **Implemented commands, queries, and events**  
✅ **Added business logic handlers**  
✅ **Enabled logging and metrics**  
✅ **Handled events automatically**  
✅ **Learned the core CQRS patterns**

## 🚀 **Next Steps**

Now that you have a working CQRS system, explore these advanced features:

1. **[Auto-Registration](./auto-registration.md)** - Automatically register handlers with dependency injection
2. **[Validation](./handlers-validators.md)** - Add command validation
3. **[Decorators](./decorators.md)** - Add cross-cutting concerns like metrics and timeouts
4. **[Performance](./performance.md)** - Optimize for production workloads
5. **[Testing](./testing.md)** - Learn testing strategies

## 💡 **Key Concepts You've Learned**

- **Commands** - Operations that change state
- **Queries** - Operations that read data  
- **Events** - Notifications about what happened
- **Handlers** - Business logic implementation
- **Manager** - Central coordinator for all operations
- **Context** - Request lifecycle and cancellation support

## 🎉 **Congratulations!**

You now have a solid foundation in CQRS with go-cqrs. The patterns you've learned here scale from simple applications to complex enterprise systems.

**Ready to build something amazing? 🚀** 