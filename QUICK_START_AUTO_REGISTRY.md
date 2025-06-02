# Quick Start: Auto-Registration for go-cqrs

## 🎯 **TL;DR - Choose Your Approach**

| Your Situation | Recommended Approach | Setup Time |
|----------------|---------------------|------------|
| **Simple handlers, max performance** | Instance Registration | 5 min |
| **Complex dependencies, modern app** | Reflection + DI | 10 min |
| **Lazy loading, factory patterns** | Factory Pattern | 15 min |
| **Large codebase, auto-discovery** | Interface Discovery | 20 min |

---

## 🚀 **Fastest Setup: Instance Registration**

**Perfect for:** Small apps, simple handlers, maximum performance

```go
func main() {
    // 1. Setup manager
    cqrs.SetManager(cqrs.NewCqrsManager().AddLoggingDecorator())
    
    // 2. Create handlers with dependencies
    userRepo := &InMemoryUserRepository{}
    handlers := []any{
        &CreateUserCommandHandler{UserRepo: userRepo},
        &GetUserQueryHandler{UserRepo: userRepo},
        &UserCreatedEventHandler{},
        &CreateUserValidator{},
    }
    
    // 3. Auto-register
    result := cqrs.AutoRegisterHandlers(handlers...)
    
    // 4. Done! 
    fmt.Printf("✅ Registered %d handlers\n", result.RegisteredHandlers)
}
```

**Pros:** Lightning fast, minimal memory, simple  
**Cons:** Manual dependency wiring

---

## 🏆 **Recommended: Reflection + Dependency Injection**

**Perfect for:** Most applications, clean code, automatic dependency injection

```go
func main() {
    // 1. Setup CQRS
    cqrs.SetManager(cqrs.NewCqrsManager().AddLoggingDecorator())
    
    // 2. Setup DI container
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, &InMemoryUserRepository{})
    cqrs.Register[EmailService](container, &ConsoleEmailService{})
    
    // 3. Auto-register with DI
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserCommandHandler{}, // Dependencies auto-injected
        &GetUserQueryHandler{},      // Dependencies auto-injected
        &UserCreatedEventHandler{},  // Dependencies auto-injected
        &CreateUserValidator{},
    )
    
    // 4. Beautiful error reporting
    if len(result.Errors) > 0 {
        for _, detail := range result.Details {
            fmt.Println(detail) // "❌ Failed to register Handler: dependency UserRepo not found"
        }
    }
}

// Handler with auto-injection
type CreateUserCommandHandler struct {
    UserRepo     UserRepository `inject:""`  // ✨ Auto-injected
    EmailService EmailService   `inject:""`  // ✨ Auto-injected
}
```

**Pros:** Clean code, auto-injection, great errors, flexible  
**Cons:** Slightly more setup

---

## 🏭 **Advanced: Factory Pattern**

**Perfect for:** Complex initialization, lazy loading, singletons

```go
func main() {
    cqrs.SetManager(cqrs.NewCqrsManager().AddLoggingDecorator())
    
    // Setup container with factories
    container := cqrs.NewSimpleContainer()
    
    // Singleton factory (created once, reused)
    cqrs.RegisterSingletonFunc[UserRepository](container, func() UserRepository {
        log.Println("🏭 Creating UserRepository singleton")
        return &PostgreSQLUserRepository{ConnectionString: getDbConnection()}
    })
    
    // Transient factory (new instance each time)
    cqrs.RegisterFunc[EmailService](container, func() EmailService {
        return &SMTPEmailService{Host: getEmailHost()}
    })
    
    // Auto-register with factories
    autoRegistry := cqrs.NewAutoRegistry(cqrs.GetManager()).SetDependencyProvider(container)
    result := autoRegistry.RegisterHandlerInstances(
        &CreateUserCommandHandler{},
        &GetUserQueryHandler{},
    )
    
    fmt.Printf("✅ Factory-based registration: %d handlers\n", result.RegisteredHandlers)
}
```

**Pros:** Lazy loading, lifecycle control, complex initialization  
**Cons:** More abstraction layers

---

## 🔧 **Expert: Interface Discovery**

**Perfect for:** Large codebases, microservices, plugin architectures

```go
func main() {
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator().AddMetricsDecorator()
    cqrs.SetManager(manager)
    
    // Advanced container with interface mapping
    container := cqrs.NewSimpleContainer()
    container.RegisterInstance(&PostgreSQLUserRepository{})
    container.RegisterInstance(&RedisCache{})
    container.RegisterInstance(&SMTPEmailService{})
    
    // Auto-discover and register entire packages
    result := cqrs.AutoRegisterFromPackage(
        handlers.UserHandlers{},
        handlers.ProductHandlers{},
        handlers.NotificationHandlers{},
    )
    
    // Comprehensive metrics
    log.Printf("🎉 Discovered and registered %d handlers across %d packages", 
        result.RegisteredHandlers, len(result.Details))
}
```

**Pros:** Minimal config, auto-discovery, scales to large apps  
**Cons:** Higher complexity, requires good architecture

---

## 🧪 **Testing Made Easy**

```go
func TestMyHandlers(t *testing.T) {
    // Setup test manager
    cqrs.SetManager(cqrs.NewCqrsManager())
    
    // Mock dependencies
    mockRepo := &MockUserRepository{}
    mockEmail := &MockEmailService{}
    
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, mockRepo)
    cqrs.Register[EmailService](container, mockEmail)
    
    // Register handlers for testing
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserCommandHandler{},
    )
    
    // Test command execution
    err := cqrs.ExecuteCommand(context.Background(), &CreateUserCommand{
        Name: "Test User", Email: "test@example.com",
    })
    
    assert.NoError(t, err)
    assert.True(t, mockRepo.SaveCalled)
}
```

---

## 🚦 **Migration from Code Generation**

**Before (Code Generation):**
```go
//go:generate gen-handler-registry
func main() {
    registerHandlers() // Generated at build time
}
```

**After (Runtime Auto-Registration):**
```go
func main() {
    cqrs.SetManager(cqrs.NewCqrsManager().AddLoggingDecorator())
    
    // Choose your approach:
    
    // Option 1: Simple
    result := cqrs.AutoRegisterHandlers(
        &CreateUserCommandHandler{UserRepo: userRepo},
        &GetUserQueryHandler{UserRepo: userRepo},
    )
    
    // Option 2: With DI (recommended)
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, &InMemoryUserRepository{})
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserCommandHandler{}, // Dependencies auto-injected
        &GetUserQueryHandler{},
    )
    
    // Handle results
    if len(result.Errors) > 0 {
        log.Fatalf("Registration failed: %v", result.Errors)
    }
    log.Printf("✅ Successfully registered %d handlers", result.RegisteredHandlers)
}
```

---

## 📊 **Performance Comparison**

| Approach | Memory/Handler | Registration Time | Runtime Overhead |
|----------|---------------|-------------------|------------------|
| Instance Registration | ~100 bytes | ~50 μs | None |
| Reflection + DI | ~200 bytes | ~150 μs | Minimal |
| Factory Pattern | ~150 bytes | ~100 μs | Lazy loading |
| Interface Discovery | ~250 bytes | ~300 μs | Discovery cost |

---

## 🎯 **Decision Matrix**

**Use Instance Registration if:**
- ✅ You have < 20 handlers
- ✅ Simple dependencies
- ✅ Need maximum performance
- ✅ Prefer explicit control

**Use Reflection + DI if:**
- ✅ You have 20-200 handlers
- ✅ Complex dependencies
- ✅ Want clean, maintainable code
- ✅ Need good error reporting

**Use Factory Pattern if:**
- ✅ Expensive-to-create dependencies
- ✅ Need lazy initialization
- ✅ Multiple lifecycle patterns
- ✅ Integration with existing DI

**Use Interface Discovery if:**
- ✅ You have 200+ handlers
- ✅ Multiple packages/modules
- ✅ Plugin architecture
- ✅ Microservices

---

## 🚀 **Get Started in 30 Seconds**

1. **Copy this into your `main.go`:**
```go
func main() {
    // Setup
    cqrs.SetManager(cqrs.NewCqrsManager().AddLoggingDecorator())
    container := cqrs.NewSimpleContainer()
    
    // Register dependencies
    cqrs.Register[UserRepository](container, &InMemoryUserRepository{})
    
    // Auto-register handlers
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserCommandHandler{},
        &GetUserQueryHandler{},
    )
    
    fmt.Printf("✅ Registered %d handlers\n", result.RegisteredHandlers)
}
```

2. **Add `inject:""` tags to your handler fields:**
```go
type CreateUserCommandHandler struct {
    UserRepo UserRepository `inject:""`
}
```

3. **Run your app and see the magic! 🎉**

---

**All approaches are runtime, dependency-aware, have excellent error logging, and use minimal resources. Choose the one that fits your app's complexity and team preferences!** 