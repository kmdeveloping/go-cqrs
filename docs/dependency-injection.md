# Dependency Injection

Clean dependency management with go-cqrs using the built-in container and auto-registration features.

## 📚 Table of Contents

- [Why Dependency Injection?](#why-dependency-injection)
- [Simple Container](#simple-container)
- [Handler Dependencies](#handler-dependencies)
- [Auto-Registration with DI](#auto-registration-with-di)
- [Lifecycle Management](#lifecycle-management)
- [Testing with DI](#testing-with-di)
- [Advanced Patterns](#advanced-patterns)
- [Best Practices](#best-practices)

## 🎯 Why Dependency Injection?

Dependency injection provides:

- **Testability**: Easy to mock dependencies for unit tests
- **Flexibility**: Swap implementations without changing code
- **Maintainability**: Clear dependencies and loose coupling
- **Configuration**: Different setups for dev/test/prod environments

```go
// ❌ Hard dependencies - difficult to test
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    db := sql.Open("postgres", "host=localhost...") // Hard-coded
    userRepo := &PostgreSQLUserRepository{db: db}   // Hard-coded
    return userRepo.Save(ctx, user)
}

// ✅ Dependency injection - easy to test and configure
type CreateUserHandler struct {
    userRepo UserRepository // Interface - can be mocked
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    return h.userRepo.Save(ctx, user) // Uses injected dependency
}
```

## 🏗️ Simple Container

go-cqrs provides a built-in dependency injection container:

### Basic Container Usage

```go
import "github.com/kmdeveloping/go-cqrs/cqrs"

// Create container
container := cqrs.NewSimpleContainer()

// Register dependencies
cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{})
cqrs.Register[EmailService](container, &SMTPEmailService{})
cqrs.Register[Logger](container, &StructuredLogger{})

// Resolve dependencies
userRepo := cqrs.Resolve[UserRepository](container)
emailService := cqrs.Resolve[EmailService](container)
```

### Interface-Based Registration

```go
// Define interfaces
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
}

type EmailService interface {
    SendWelcome(ctx context.Context, email, name string) error
    SendPasswordReset(ctx context.Context, email, token string) error
}

// Register implementations
cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{
    connectionString: "postgres://localhost/mydb",
})

cqrs.Register[EmailService](container, &SMTPEmailService{
    host:     "smtp.example.com",
    port:     587,
    username: "noreply@example.com",
    password: "smtp-password",
})
```

## 🎯 Handler Dependencies

### Manual Dependency Injection

```go
type CreateUserHandler struct {
    userRepo     UserRepository
    emailService EmailService
    logger       Logger
}

func NewCreateUserHandler(
    userRepo UserRepository,
    emailService EmailService,
    logger Logger,
) *CreateUserHandler {
    return &CreateUserHandler{
        userRepo:     userRepo,
        emailService: emailService,
        logger:       logger,
    }
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Use injected dependencies
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        h.logger.Error("Failed to save user", err)
        return err
    }
    
    return h.emailService.SendWelcome(ctx, user.Email, user.Name)
}

// Wire up manually
func setupHandlers(container *cqrs.SimpleContainer) {
    userRepo := cqrs.Resolve[UserRepository](container)
    emailService := cqrs.Resolve[EmailService](container)
    logger := cqrs.Resolve[Logger](container)
    
    createUserHandler := NewCreateUserHandler(userRepo, emailService, logger)
    cqrs.RegisterCommandHandler(createUserHandler)
}
```

### Tag-Based Injection

```go
type CreateUserHandler struct {
    UserRepo     UserRepository `inject:""`
    EmailService EmailService   `inject:""`
    Logger       Logger         `inject:""`
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Dependencies automatically injected
    user := &User{Name: cmd.Name, Email: cmd.Email}
    
    if err := h.UserRepo.Save(ctx, user); err != nil {
        h.Logger.Error("Failed to save user", err)
        return err
    }
    
    return h.EmailService.SendWelcome(ctx, user.Email, user.Name)
}
```

## 🚀 Auto-Registration with DI

The most powerful feature - automatic dependency injection during registration:

### Basic Auto-Registration

```go
func main() {
    // Setup CQRS
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Setup container with dependencies
    container := cqrs.NewSimpleContainer()
    
    // Register services
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{
        connectionString: getDBConnectionString(),
    })
    
    cqrs.Register[EmailService](container, &SMTPEmailService{
        host:     getEmailHost(),
        username: getEmailUser(),
        password: getEmailPass(),
    })
    
    cqrs.Register[Logger](container, &StructuredLogger{
        level: "INFO",
    })
    
    // Auto-register handlers with dependency injection
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserHandler{}, // Dependencies auto-injected
        &UpdateUserHandler{}, // Dependencies auto-injected
        &DeleteUserHandler{}, // Dependencies auto-injected
        &GetUserHandler{},    // Dependencies auto-injected
        &ListUsersHandler{},  // Dependencies auto-injected
    )
    
    fmt.Printf("✅ Registered %d handlers with auto-injected dependencies\n", 
        result.RegisteredHandlers)
}
```

### Complex Service Setup

```go
// Database setup
func setupDatabase() *sql.DB {
    db, err := sql.Open("postgres", getDatabaseURL())
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    return db
}

// Redis setup
func setupRedis() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     getRedisAddr(),
        Password: getRedisPassword(),
        DB:       0,
    })
}

// Complete application setup
func setupApplication() {
    // Core infrastructure
    db := setupDatabase()
    redisClient := setupRedis()
    
    // Container setup
    container := cqrs.NewSimpleContainer()
    
    // Register infrastructure
    cqrs.Register[*sql.DB](container, db)
    cqrs.Register[*redis.Client](container, redisClient)
    
    // Register repositories
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{DB: db})
    cqrs.Register[OrderRepository](container, &PostgreSQLOrderRepository{DB: db})
    cqrs.Register[ProductRepository](container, &PostgreSQLProductRepository{DB: db})
    
    // Register services
    cqrs.Register[EmailService](container, &SMTPEmailService{
        Host:     getEmailHost(),
        Username: getEmailUser(),
        Password: getEmailPass(),
    })
    
    cqrs.Register[PaymentService](container, &StripePaymentService{
        apiKey: getStripeAPIKey(),
    })
    
    cqrs.Register[CacheService](container, &RedisCacheService{
        client: redisClient,
        ttl:    5 * time.Minute,
    })
    
    // Auto-register all handlers
    result := cqrs.AutoRegisterWithDependencies(container,
        // User handlers
        &CreateUserHandler{},
        &UpdateUserHandler{},
        &GetUserHandler{},
        &ListUsersHandler{},
        
        // Order handlers
        &CreateOrderHandler{},
        &UpdateOrderStatusHandler{},
        &GetOrderHandler{},
        &ListOrdersHandler{},
        
        // Product handlers
        &CreateProductHandler{},
        &UpdateProductHandler{},
        &GetProductHandler{},
        &SearchProductsHandler{},
        
        // Event handlers
        &UserCreatedEventHandler{},
        &OrderPlacedEventHandler{},
        &PaymentProcessedEventHandler{},
        
        // Validators
        &CreateUserValidator{},
        &CreateOrderValidator{},
    )
    
    log.Printf("✅ Application setup complete:")
    log.Printf("   - Registered %d handlers", result.RegisteredHandlers)
    log.Printf("   - Registered %d validators", result.RegisteredValidators)
}
```

## 🔄 Lifecycle Management

### Singleton vs Transient

```go
// Singleton - single instance shared across application
cqrs.Register[DatabaseConnection](container, &PostgreSQLConnection{
    connectionString: getDBURL(),
})

// Factory function for transient instances
cqrs.RegisterFactory[HTTPClient](container, func() HTTPClient {
    return &http.Client{
        Timeout: 30 * time.Second,
    }
})

// Usage in handler
type ExternalAPIHandler struct {
    dbConn     DatabaseConnection `inject:""`     // Singleton
    httpClient HTTPClient         `inject:""`     // New instance each time
}
```

### Initialization and Cleanup

```go
type EmailService interface {
    SendEmail(ctx context.Context, to, subject, body string) error
    Close() error // Cleanup method
}

type SMTPEmailService struct {
    host     string
    port     int
    client   *smtp.Client
}

func (s *SMTPEmailService) Initialize() error {
    client, err := smtp.Dial(fmt.Sprintf("%s:%d", s.host, s.port))
    if err != nil {
        return err
    }
    s.client = client
    return nil
}

func (s *SMTPEmailService) Close() error {
    if s.client != nil {
        return s.client.Close()
    }
    return nil
}

// Register with initialization
func setupEmailService(container *cqrs.SimpleContainer) {
    emailService := &SMTPEmailService{
        host: getEmailHost(),
        port: getEmailPort(),
    }
    
    if err := emailService.Initialize(); err != nil {
        log.Fatal("Failed to initialize email service:", err)
    }
    
    cqrs.Register[EmailService](container, emailService)
    
    // Setup cleanup on application shutdown
    gracefulShutdown(func() {
        emailService.Close()
    })
}
```

## 🧪 Testing with DI

### Mock Dependencies

```go
// Mock implementations for testing
type MockUserRepository struct {
    users map[int]*User
    nextID int
}

func (m *MockUserRepository) Save(ctx context.Context, user *User) error {
    m.nextID++
    user.ID = m.nextID
    m.users[user.ID] = user
    return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    user, exists := m.users[id]
    if !exists {
        return nil, errors.New("user not found")
    }
    return user, nil
}

type MockEmailService struct {
    sentEmails []SentEmail
}

type SentEmail struct {
    To      string
    Subject string
    Body    string
}

func (m *MockEmailService) SendWelcome(ctx context.Context, email, name string) error {
    m.sentEmails = append(m.sentEmails, SentEmail{
        To:      email,
        Subject: "Welcome!",
        Body:    fmt.Sprintf("Welcome %s!", name),
    })
    return nil
}

// Test setup
func TestCreateUserHandler(t *testing.T) {
    // Setup test container
    container := cqrs.NewSimpleContainer()
    
    // Register mock dependencies
    mockUserRepo := &MockUserRepository{users: make(map[int]*User)}
    mockEmailService := &MockEmailService{}
    
    cqrs.Register[UserRepository](container, mockUserRepo)
    cqrs.Register[EmailService](container, mockEmailService)
    
    // Create handler with injected dependencies
    handler := &CreateUserHandler{}
    cqrs.InjectDependencies(container, handler)
    
    // Test the handler
    ctx := context.Background()
    cmd := &CreateUserCommand{
        Name:  "Test User",
        Email: "test@example.com",
    }
    
    err := handler.Handle(ctx, cmd)
    
    // Assertions
    assert.NoError(t, err)
    assert.Len(t, mockUserRepo.users, 1)
    assert.Len(t, mockEmailService.sentEmails, 1)
    assert.Equal(t, "test@example.com", mockEmailService.sentEmails[0].To)
}
```

### Test Utilities

```go
// Test helper for setting up container
func SetupTestContainer() *cqrs.SimpleContainer {
    container := cqrs.NewSimpleContainer()
    
    // Register test implementations
    cqrs.Register[UserRepository](container, &MockUserRepository{
        users: make(map[int]*User),
    })
    
    cqrs.Register[EmailService](container, &MockEmailService{})
    
    cqrs.Register[Logger](container, &TestLogger{})
    
    return container
}

// Integration test setup
func SetupIntegrationTestContainer() *cqrs.SimpleContainer {
    container := cqrs.NewSimpleContainer()
    
    // Use real database but test instance
    testDB := setupTestDatabase()
    
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{
        DB: testDB,
    })
    
    // Mock external services
    cqrs.Register[EmailService](container, &MockEmailService{})
    
    return container
}
```

## 🏗️ Advanced Patterns

### Conditional Registration

```go
func setupServices(container *cqrs.SimpleContainer, env string) {
    // Different implementations based on environment
    switch env {
    case "development":
        cqrs.Register[EmailService](container, &ConsoleEmailService{})
        cqrs.Register[PaymentService](container, &MockPaymentService{})
    case "testing":
        cqrs.Register[EmailService](container, &MockEmailService{})
        cqrs.Register[PaymentService](container, &MockPaymentService{})
    case "production":
        cqrs.Register[EmailService](container, &SMTPEmailService{
            host:     getEmailHost(),
            username: getEmailUser(),
            password: getEmailPass(),
        })
        cqrs.Register[PaymentService](container, &StripePaymentService{
            apiKey: getStripeAPIKey(),
        })
    }
}
```

### Decorator Pattern with DI

```go
type LoggingUserRepository struct {
    baseRepo UserRepository
    logger   Logger
}

func (r *LoggingUserRepository) Save(ctx context.Context, user *User) error {
    r.logger.Info("Saving user", map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
    })
    
    err := r.baseRepo.Save(ctx, user)
    
    if err != nil {
        r.logger.Error("Failed to save user", err)
    } else {
        r.logger.Info("User saved successfully")
    }
    
    return err
}

// Setup decorated repository
func setupRepositories(container *cqrs.SimpleContainer) {
    // Base repository
    baseRepo := &PostgreSQLUserRepository{DB: getDB()}
    
    // Decorated repository
    logger := cqrs.Resolve[Logger](container)
    decoratedRepo := &LoggingUserRepository{
        baseRepo: baseRepo,
        logger:   logger,
    }
    
    cqrs.Register[UserRepository](container, decoratedRepo)
}
```

### Configuration-Based Setup

```go
type AppConfig struct {
    Database DatabaseConfig `json:"database"`
    Email    EmailConfig    `json:"email"`
    Redis    RedisConfig    `json:"redis"`
}

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
    Username string `json:"username"`
    Password string `json:"password"`
}

func setupFromConfig(configPath string) *cqrs.SimpleContainer {
    config := loadConfig(configPath)
    container := cqrs.NewSimpleContainer()
    
    // Setup database
    db := setupDatabaseFromConfig(config.Database)
    cqrs.Register[*sql.DB](container, db)
    
    // Setup repositories
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{DB: db})
    
    // Setup services based on config
    emailService := createEmailService(config.Email)
    cqrs.Register[EmailService](container, emailService)
    
    return container
}
```

## 🎓 Best Practices

### 1. Use Interfaces

```go
// ✅ Good - depend on interfaces
type CreateUserHandler struct {
    UserRepo     UserRepository `inject:""`
    EmailService EmailService   `inject:""`
}

// ❌ Bad - depend on concrete types
type CreateUserHandler struct {
    UserRepo     *PostgreSQLUserRepository `inject:""`
    EmailService *SMTPEmailService         `inject:""`
}
```

### 2. Keep Dependencies Minimal

```go
// ✅ Good - only needed dependencies
type CreateUserHandler struct {
    UserRepo UserRepository `inject:""`
}

// ❌ Bad - too many dependencies
type CreateUserHandler struct {
    UserRepo      UserRepository   `inject:""`
    OrderRepo     OrderRepository  `inject:""`
    ProductRepo   ProductRepository `inject:""`
    EmailService  EmailService     `inject:""`
    SMS           SMSService       `inject:""`
    Analytics     AnalyticsService `inject:""`
    // ... too many
}
```

### 3. Separate Configuration

```go
// ✅ Good - configuration separate from business logic
type EmailConfig struct {
    Host     string
    Port     int
    Username string
    Password string
}

func setupEmailService(config EmailConfig) EmailService {
    return &SMTPEmailService{
        host:     config.Host,
        port:     config.Port,
        username: config.Username,
        password: config.Password,
    }
}

// ❌ Bad - configuration mixed with business logic
type SMTPEmailService struct {
    // Don't hardcode configuration
    host string // = "smtp.gmail.com"
    port int    // = 587
}
```

### 4. Lifecycle Awareness

```go
// ✅ Good - proper cleanup
type DatabaseConnection struct {
    db *sql.DB
}

func (d *DatabaseConnection) Close() error {
    return d.db.Close()
}

// Register with cleanup
func setupDatabase(container *cqrs.SimpleContainer) {
    db := &DatabaseConnection{db: openDB()}
    cqrs.Register[DatabaseConnection](container, db)
    
    // Ensure cleanup on shutdown
    onShutdown(func() {
        db.Close()
    })
}
```

## 🚀 Next Steps

1. **Learn Auto-Registration**: [Auto-Registration Guide](./auto-registration.md)
2. **Add Decorators**: [Decorators & Middleware](./decorators.md)
3. **Testing Strategies**: [Testing Strategies](./testing.md)
4. **Production Setup**: [Production Readiness](./production-ready.md)

---

**Ready for advanced features? Continue with [Auto-Registration Guide](./auto-registration.md)! 🚀** 