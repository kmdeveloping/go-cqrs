# Installation & Setup Guide

## 📦 Installation

### Prerequisites

- Go 1.19 or later
- Go modules enabled

### Install via Go Modules

```bash
go get github.com/kmdeveloping/go-cqrs
```

### Verify Installation

Create a simple test file to verify installation:

```go
// test_installation.go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
)

type TestCommand struct {
    command.Base
    Message string
}

type TestHandler struct{}

func (h *TestHandler) Handle(ctx context.Context, cmd *TestCommand) error {
    fmt.Printf("✅ Installation working! Message: %s\n", cmd.Message)
    return nil
}

func main() {
    manager := cqrs.NewCqrsManager()
    if err := cqrs.SetManager(manager); err != nil {
        log.Fatal(err)
    }
    
    cqrs.RegisterCommandHandler(&TestHandler{})
    
    ctx := context.Background()
    err := cqrs.ExecuteCommand(ctx, &TestCommand{Message: "Hello go-cqrs!"})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("🚀 go-cqrs is ready to use!")
}
```

Run the test:

```bash
go run test_installation.go
```

## 🏗️ Project Setup

### Basic Project Structure

```
your-project/
├── main.go
├── go.mod
├── go.sum
├── commands/
│   ├── user_commands.go
│   └── handlers/
│       └── user_handlers.go
├── queries/
│   ├── user_queries.go
│   └── handlers/
│       └── user_handlers.go
├── events/
│   ├── user_events.go
│   └── handlers/
│       └── user_handlers.go
└── validators/
    └── user_validators.go
```

### Initialize Go Module

```bash
mkdir my-cqrs-app
cd my-cqrs-app
go mod init my-cqrs-app
go get github.com/kmdeveloping/go-cqrs
```

### Basic Setup

Create your `main.go`:

```go
package main

import (
    "context"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

func main() {
    // Initialize CQRS manager
    manager := cqrs.NewCqrsManager()
    
    // Add optional decorators
    manager.AddLoggingDecorator()
    
    // Set as global manager
    if err := cqrs.SetManager(manager); err != nil {
        log.Fatal("Failed to set CQRS manager:", err)
    }
    
    // Register your handlers here
    registerHandlers()
    
    // Start your application
    startApplication()
}

func registerHandlers() {
    // Register command handlers
    // cqrs.RegisterCommandHandler(&YourCommandHandler{})
    
    // Register query handlers  
    // cqrs.RegisterQueryHandler(&YourQueryHandler{})
    
    // Register event handlers
    // cqrs.RegisterEventHandler(&YourEventHandler{})
}

func startApplication() {
    ctx := context.Background()
    
    // Your application logic here
    log.Println("🚀 CQRS application started!")
    
    // Keep application running
    select {}
}
```

## 🔧 Configuration Options

### Manager Configuration

```go
manager := cqrs.NewCqrsManager()

// Add logging decorator (optional)
manager.AddLoggingDecorator()

// Configure with custom options (if available)
// manager.WithTimeout(30 * time.Second)
// manager.WithMaxRetries(3)
```

### Auto-Registration Setup

```go
// Setup dependency injection container
container := cqrs.NewSimpleContainer()

// Register dependencies
cqrs.Register[DatabaseConnection](container, &PostgreSQLConnection{})
cqrs.Register[Logger](container, &StructuredLogger{})

// Auto-register handlers with dependencies
result := cqrs.AutoRegisterWithDependencies(container,
    &CreateUserHandler{},
    &GetUserHandler{},
    &UserCreatedHandler{},
)

fmt.Printf("✅ Registered %d handlers\n", result.RegisteredHandlers)
```

## 🧪 Development Environment

### Recommended IDE Setup

**VS Code Extensions:**
- Go extension
- Go Test Explorer
- GitLens

**GoLand/IntelliJ:**
- Native Go support
- Database tools (for data persistence)

### Environment Variables

Create a `.env` file for configuration:

```env
# Development settings
GO_ENV=development
LOG_LEVEL=debug

# Database settings (if using)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_dev
DB_USER=dev_user
DB_PASS=dev_pass

# Service settings
HTTP_PORT=8080
GRPC_PORT=9090
```

### Development Dependencies

Add common development dependencies:

```bash
# Testing
go get -t github.com/stretchr/testify/assert
go get -t github.com/stretchr/testify/mock

# Configuration
go get github.com/joho/godotenv

# Logging
go get github.com/sirupsen/logrus

# HTTP server (if building web API)
go get github.com/gin-gonic/gin
```

## 🚀 Next Steps

1. **Learn the Fundamentals:** Read [CQRS Fundamentals](./cqrs-fundamentals.md)
2. **See Examples:** Check [Basic Examples](./basic-examples.md)
3. **Build Your First Handler:** Follow [Commands, Queries & Events](./commands-queries-events.md)
4. **Add Testing:** See [Testing Strategies](./testing.md)

## 🔍 Troubleshooting

### Common Issues

**Go version compatibility:**
```bash
go version
# Should be 1.19 or later
```

**Module not found:**
```bash
go mod tidy
go clean -modcache
go get github.com/kmdeveloping/go-cqrs
```

**Import issues:**
```go
// Make sure imports are correct
import (
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/query"
    "github.com/kmdeveloping/go-cqrs/event"
)
```

## 💡 Tips

- Use `go mod tidy` regularly to keep dependencies clean
- Set up proper project structure from the beginning
- Use interfaces for better testability
- Consider using dependency injection for complex applications

---

**Ready to build? Check out the [Quick Start Guide](./quick-start.md)! 🚀** 