# Testing Strategies

Comprehensive testing approaches for go-cqrs applications including unit tests, integration tests, and test utilities.

## 📚 Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Unit Testing](#unit-testing)
- [Integration Testing](#integration-testing)
- [Test Utilities](#test-utilities)
- [Mocking Strategies](#mocking-strategies)
- [Event Testing](#event-testing)
- [Performance Testing](#performance-testing)
- [Best Practices](#best-practices)

## 🎯 Testing Philosophy

CQRS applications benefit from clear separation of concerns, making them highly testable:

- **Commands**: Test business logic and side effects
- **Queries**: Test data retrieval and formatting
- **Events**: Test event handling and state changes
- **Handlers**: Test individual responsibilities
- **Validators**: Test validation rules

```go
// Each component can be tested in isolation
func TestCreateUserCommand(t *testing.T) {
    // Test command validation
}

func TestCreateUserHandler(t *testing.T) {
    // Test business logic
}

func TestUserCreatedEvent(t *testing.T) {
    // Test event handling
}
```

## 🔬 Unit Testing

### Command Handler Testing

```go
func TestCreateUserHandler(t *testing.T) {
    tests := []struct {
        name        string
        command     *CreateUserCommand
        setupMocks  func(*MockUserRepository, *MockEmailService)
        expectError bool
        validate    func(*testing.T, *MockUserRepository, *MockEmailService)
    }{
        {
            name: "successful user creation",
            command: &CreateUserCommand{
                Name:  "John Doe",
                Email: "john@example.com",
            },
            setupMocks: func(userRepo *MockUserRepository, emailService *MockEmailService) {
                userRepo.On("Save", mock.Anything, mock.AnythingOfType("*User")).Return(nil)
                emailService.On("SendWelcome", mock.Anything, "john@example.com", "John Doe").Return(nil)
            },
            expectError: false,
            validate: func(t *testing.T, userRepo *MockUserRepository, emailService *MockEmailService) {
                userRepo.AssertExpectations(t)
                emailService.AssertExpectations(t)
            },
        },
        {
            name: "database error",
            command: &CreateUserCommand{
                Name:  "John Doe",
                Email: "john@example.com",
            },
            setupMocks: func(userRepo *MockUserRepository, emailService *MockEmailService) {
                userRepo.On("Save", mock.Anything, mock.AnythingOfType("*User")).Return(errors.New("database error"))
            },
            expectError: true,
            validate: func(t *testing.T, userRepo *MockUserRepository, emailService *MockEmailService) {
                userRepo.AssertExpectations(t)
                emailService.AssertNotCalled(t, "SendWelcome")
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            userRepo := &MockUserRepository{}
            emailService := &MockEmailService{}
            logger := &MockLogger{}
            
            tt.setupMocks(userRepo, emailService)
            
            // Create handler
            handler := &CreateUserHandler{
                userRepo:     userRepo,
                emailService: emailService,
                logger:       logger,
            }
            
            // Execute
            ctx := context.Background()
            err := handler.Handle(ctx, tt.command)
            
            // Assert
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            // Custom validation
            tt.validate(t, userRepo, emailService)
        })
    }
}
```

### Query Handler Testing

```go
func TestGetUserHandler(t *testing.T) {
    tests := []struct {
        name         string
        query        GetUserQuery
        setupMock    func(*MockUserRepository)
        expectedUser *User
        expectError  bool
    }{
        {
            name: "user found",
            query: GetUserQuery{UserID: 1},
            setupMock: func(userRepo *MockUserRepository) {
                user := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
                userRepo.On("GetByID", mock.Anything, 1).Return(user, nil)
            },
            expectedUser: &User{ID: 1, Name: "John Doe", Email: "john@example.com"},
            expectError:  false,
        },
        {
            name: "user not found",
            query: GetUserQuery{UserID: 999},
            setupMock: func(userRepo *MockUserRepository) {
                userRepo.On("GetByID", mock.Anything, 999).Return(nil, ErrUserNotFound)
            },
            expectedUser: nil,
            expectError:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            userRepo := &MockUserRepository{}
            tt.setupMock(userRepo)
            
            handler := &GetUserHandler{userRepo: userRepo}
            
            // Execute
            ctx := context.Background()
            user, err := handler.Handle(ctx, tt.query)
            
            // Assert
            if tt.expectError {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedUser, user)
            }
            
            userRepo.AssertExpectations(t)
        })
    }
}
```

### Validator Testing

```go
func TestCreateUserValidator(t *testing.T) {
    tests := []struct {
        name        string
        command     *CreateUserCommand
        setupMock   func(*MockUserRepository)
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid command",
            command: &CreateUserCommand{
                Name:     "John Doe",
                Email:    "john@example.com",
                Password: "password123",
            },
            setupMock: func(userRepo *MockUserRepository) {
                userRepo.On("ExistsByEmail", mock.Anything, "john@example.com").Return(false, nil)
            },
            expectError: false,
        },
        {
            name: "empty name",
            command: &CreateUserCommand{
                Name:     "",
                Email:    "john@example.com",
                Password: "password123",
            },
            setupMock:   func(userRepo *MockUserRepository) {},
            expectError: true,
            errorMsg:    "name is required",
        },
        {
            name: "invalid email",
            command: &CreateUserCommand{
                Name:     "John Doe",
                Email:    "invalid-email",
                Password: "password123",
            },
            setupMock:   func(userRepo *MockUserRepository) {},
            expectError: true,
            errorMsg:    "invalid email format",
        },
        {
            name: "email already exists",
            command: &CreateUserCommand{
                Name:     "John Doe",
                Email:    "john@example.com",
                Password: "password123",
            },
            setupMock: func(userRepo *MockUserRepository) {
                userRepo.On("ExistsByEmail", mock.Anything, "john@example.com").Return(true, nil)
            },
            expectError: true,
            errorMsg:    "email already exists",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            userRepo := &MockUserRepository{}
            tt.setupMock(userRepo)
            
            validator := &CreateUserValidator{userRepo: userRepo}
            
            // Execute
            ctx := context.Background()
            err := validator.Validate(ctx, tt.command)
            
            // Assert
            if tt.expectError {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
            } else {
                assert.NoError(t, err)
            }
            
            userRepo.AssertExpectations(t)
        })
    }
}
```

## 🔗 Integration Testing

### End-to-End Command Testing

```go
func TestCreateUserIntegration(t *testing.T) {
    // Setup test environment
    defer cqrs.ResetManager()
    
    // Setup test database
    testDB := setupTestDatabase(t)
    defer testDB.Close()
    
    // Setup container with real dependencies
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{DB: testDB})
    cqrs.Register[EmailService](container, &MockEmailService{})
    cqrs.Register[Logger](container, &TestLogger{})
    
    // Setup CQRS manager
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Auto-register handlers
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserHandler{},
        &GetUserHandler{},
        &CreateUserValidator{},
    )
    
    assert.Equal(t, 2, result.RegisteredHandlers)
    assert.Equal(t, 1, result.RegisteredValidators)
    
    // Test command execution
    ctx := context.Background()
    
    // Execute create command
    createCmd := &CreateUserCommand{
        Name:     "Integration Test User",
        Email:    "integration@example.com",
        Password: "password123",
    }
    
    err := cqrs.ExecuteCommand(ctx, createCmd)
    assert.NoError(t, err)
    
    // Verify user was created by querying
    getUserQuery := GetUserQuery{UserID: 1}
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, getUserQuery)
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "Integration Test User", user.Name)
    assert.Equal(t, "integration@example.com", user.Email)
}
```

### Event Integration Testing

```go
func TestUserCreatedEventIntegration(t *testing.T) {
    defer cqrs.ResetManager()
    
    // Setup
    container := cqrs.NewSimpleContainer()
    mockEmailService := &MockEmailService{}
    mockAnalytics := &MockAnalyticsService{}
    
    cqrs.Register[EmailService](container, mockEmailService)
    cqrs.Register[AnalyticsService](container, mockAnalytics)
    
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register event handlers
    cqrs.AutoRegisterWithDependencies(container,
        &UserCreatedEmailHandler{},
        &UserCreatedAnalyticsHandler{},
    )
    
    // Setup expectations
    mockEmailService.On("SendWelcome", mock.Anything, "test@example.com", "Test User").Return(nil)
    mockAnalytics.On("TrackUserRegistration", mock.Anything, 123).Return(nil)
    
    // Publish event
    ctx := context.Background()
    event := UserCreatedEvent{
        Base: event.Base{
            ExecutionTime:  time.Now(),
            CorrelationUid: uuid.New(),
        },
        UserID: 123,
        Name:   "Test User",
        Email:  "test@example.com",
    }
    
    err := cqrs.PublishEvent(ctx, event)
    assert.NoError(t, err)
    
    // Allow event processing
    time.Sleep(100 * time.Millisecond)
    
    // Verify expectations
    mockEmailService.AssertExpectations(t)
    mockAnalytics.AssertExpectations(t)
}
```

## 🛠️ Test Utilities

### Test Container Setup

```go
func SetupTestContainer() *cqrs.SimpleContainer {
    container := cqrs.NewSimpleContainer()
    
    // Register test implementations
    cqrs.Register[UserRepository](container, &MockUserRepository{})
    cqrs.Register[EmailService](container, &MockEmailService{})
    cqrs.Register[Logger](container, &TestLogger{})
    
    return container
}

func SetupIntegrationTestContainer(t *testing.T) *cqrs.SimpleContainer {
    container := cqrs.NewSimpleContainer()
    
    // Real database for integration tests
    testDB := setupTestDatabase(t)
    t.Cleanup(func() { testDB.Close() })
    
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{DB: testDB})
    
    // Mock external services
    cqrs.Register[EmailService](container, &MockEmailService{})
    cqrs.Register[PaymentService](container, &MockPaymentService{})
    
    return container
}
```

### Test Database Setup

```go
func setupTestDatabase(t *testing.T) *sql.DB {
    // Use test database
    dsn := "postgres://test:test@localhost/test_db?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    // Clean database before each test
    cleanTestDatabase(t, db)
    
    return db
}

func cleanTestDatabase(t *testing.T, db *sql.DB) {
    tables := []string{"users", "orders", "products"}
    
    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
        require.NoError(t, err)
    }
    
    // Reset sequences
    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table))
        require.NoError(t, err)
    }
}
```

### Test Helpers

```go
// Helper for creating test users
func CreateTestUser(t *testing.T, userRepo UserRepository, name, email string) *User {
    user := &User{
        Name:      name,
        Email:     email,
        CreatedAt: time.Now(),
    }
    
    ctx := context.Background()
    err := userRepo.Save(ctx, user)
    require.NoError(t, err)
    
    return user
}

// Helper for creating test commands
func NewTestCreateUserCommand(name, email string) *CreateUserCommand {
    return &CreateUserCommand{
        Name:     name,
        Email:    email,
        Password: "test-password-123",
    }
}

// Helper for asserting events
func AssertEventPublished(t *testing.T, eventBus *MockEventBus, eventType interface{}) {
    found := false
    for _, event := range eventBus.PublishedEvents {
        if reflect.TypeOf(event) == reflect.TypeOf(eventType) {
            found = true
            break
        }
    }
    assert.True(t, found, "Expected event %T was not published", eventType)
}
```

## 🎭 Mocking Strategies

### Repository Mocks

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Save(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
    args := m.Called(ctx, email)
    return args.Bool(0), args.Error(1)
}
```

### Service Mocks

```go
type MockEmailService struct {
    mock.Mock
    sentEmails []SentEmail
}

type SentEmail struct {
    To      string
    Subject string
    Body    string
    SentAt  time.Time
}

func (m *MockEmailService) SendWelcome(ctx context.Context, email, name string) error {
    args := m.Called(ctx, email, name)
    
    if args.Error(0) == nil {
        m.sentEmails = append(m.sentEmails, SentEmail{
            To:      email,
            Subject: "Welcome!",
            Body:    fmt.Sprintf("Welcome %s!", name),
            SentAt:  time.Now(),
        })
    }
    
    return args.Error(0)
}

func (m *MockEmailService) GetSentEmails() []SentEmail {
    return m.sentEmails
}

func (m *MockEmailService) ClearSentEmails() {
    m.sentEmails = nil
}
```

### In-Memory Test Implementations

```go
type InMemoryUserRepository struct {
    users  map[int]*User
    nextID int
    mu     sync.RWMutex
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
    return &InMemoryUserRepository{
        users:  make(map[int]*User),
        nextID: 1,
    }
}

func (r *InMemoryUserRepository) Save(ctx context.Context, user *User) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if user.ID == 0 {
        user.ID = r.nextID
        r.nextID++
    }
    
    user.UpdatedAt = time.Now()
    r.users[user.ID] = user
    
    return nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    user, exists := r.users[id]
    if !exists {
        return nil, ErrUserNotFound
    }
    
    return user, nil
}
```

## 📢 Event Testing

### Event Handler Testing

```go
func TestUserCreatedEventHandler(t *testing.T) {
    tests := []struct {
        name       string
        event      UserCreatedEvent
        setupMocks func(*MockEmailService, *MockAnalyticsService)
        expectError bool
    }{
        {
            name: "successful event handling",
            event: UserCreatedEvent{
                UserID: 123,
                Name:   "John Doe",
                Email:  "john@example.com",
            },
            setupMocks: func(emailService *MockEmailService, analytics *MockAnalyticsService) {
                emailService.On("SendWelcome", mock.Anything, "john@example.com", "John Doe").Return(nil)
                analytics.On("TrackUserRegistration", mock.Anything, 123).Return(nil)
            },
            expectError: false,
        },
        {
            name: "email service failure",
            event: UserCreatedEvent{
                UserID: 123,
                Name:   "John Doe",
                Email:  "john@example.com",
            },
            setupMocks: func(emailService *MockEmailService, analytics *MockAnalyticsService) {
                emailService.On("SendWelcome", mock.Anything, "john@example.com", "John Doe").Return(errors.New("email error"))
                // Analytics should still be called
                analytics.On("TrackUserRegistration", mock.Anything, 123).Return(nil)
            },
            expectError: false, // Email failure shouldn't fail the event
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            emailService := &MockEmailService{}
            analytics := &MockAnalyticsService{}
            logger := &MockLogger{}
            
            tt.setupMocks(emailService, analytics)
            
            handler := &UserCreatedEventHandler{
                emailService: emailService,
                analytics:    analytics,
                logger:       logger,
            }
            
            // Execute
            ctx := context.Background()
            err := handler.Handle(ctx, tt.event)
            
            // Assert
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            emailService.AssertExpectations(t)
            analytics.AssertExpectations(t)
        })
    }
}
```

### Event Publishing Testing

```go
func TestEventPublishing(t *testing.T) {
    defer cqrs.ResetManager()
    
    // Setup event bus mock
    eventBus := &MockEventBus{}
    manager := cqrs.NewCqrsManager()
    manager.SetEventBus(eventBus)
    cqrs.SetManager(manager)
    
    // Setup expectation
    expectedEvent := UserCreatedEvent{
        UserID: 123,
        Name:   "Test User",
        Email:  "test@example.com",
    }
    
    eventBus.On("Publish", mock.Anything, mock.MatchedBy(func(event UserCreatedEvent) bool {
        return event.UserID == 123 && event.Name == "Test User"
    })).Return(nil)
    
    // Publish event
    ctx := context.Background()
    err := cqrs.PublishEvent(ctx, expectedEvent)
    
    // Assert
    assert.NoError(t, err)
    eventBus.AssertExpectations(t)
}
```

## ⚡ Performance Testing

### Benchmark Tests

```go
func BenchmarkCreateUserCommand(b *testing.B) {
    // Setup
    container := SetupTestContainer()
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    cqrs.AutoRegisterWithDependencies(container, &CreateUserHandler{})
    
    cmd := &CreateUserCommand{
        Name:     "Benchmark User",
        Email:    "benchmark@example.com",
        Password: "password123",
    }
    
    ctx := context.Background()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        cmd.Email = fmt.Sprintf("user%d@example.com", i)
        err := cqrs.ExecuteCommand(ctx, cmd)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGetUserQuery(b *testing.B) {
    // Setup with pre-populated data
    container := SetupTestContainer()
    userRepo := container.Resolve[UserRepository]().(*MockUserRepository)
    
    // Pre-populate users
    for i := 1; i <= 1000; i++ {
        user := &User{ID: i, Name: fmt.Sprintf("User %d", i)}
        userRepo.users[i] = user
    }
    
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    cqrs.AutoRegisterWithDependencies(container, &GetUserHandler{})
    
    query := GetUserQuery{UserID: 1}
    ctx := context.Background()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        query.UserID = (i % 1000) + 1
        _, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, query)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Load Testing

```go
func TestConcurrentCommands(t *testing.T) {
    defer cqrs.ResetManager()
    
    // Setup
    container := SetupTestContainer()
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    cqrs.AutoRegisterWithDependencies(container, &CreateUserHandler{})
    
    const numGoroutines = 100
    const commandsPerGoroutine = 10
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines*commandsPerGoroutine)
    
    // Launch concurrent commands
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(goroutineID int) {
            defer wg.Done()
            
            for j := 0; j < commandsPerGoroutine; j++ {
                cmd := &CreateUserCommand{
                    Name:     fmt.Sprintf("User %d-%d", goroutineID, j),
                    Email:    fmt.Sprintf("user%d-%d@example.com", goroutineID, j),
                    Password: "password123",
                }
                
                ctx := context.Background()
                if err := cqrs.ExecuteCommand(ctx, cmd); err != nil {
                    errors <- err
                    return
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Error("Command failed:", err)
    }
}
```

## 🎓 Best Practices

### 1. Test Structure

```go
// ✅ Good - clear test structure
func TestCreateUserHandler(t *testing.T) {
    // Arrange
    userRepo := &MockUserRepository{}
    emailService := &MockEmailService{}
    handler := &CreateUserHandler{userRepo: userRepo, emailService: emailService}
    
    userRepo.On("Save", mock.Anything, mock.AnythingOfType("*User")).Return(nil)
    emailService.On("SendWelcome", mock.Anything, "test@example.com", "Test User").Return(nil)
    
    // Act
    ctx := context.Background()
    cmd := &CreateUserCommand{Name: "Test User", Email: "test@example.com"}
    err := handler.Handle(ctx, cmd)
    
    // Assert
    assert.NoError(t, err)
    userRepo.AssertExpectations(t)
    emailService.AssertExpectations(t)
}
```

### 2. Isolation

```go
// ✅ Good - each test is isolated
func TestEachHandlerSeparately(t *testing.T) {
    t.Run("CreateUser", func(t *testing.T) {
        // Test only CreateUserHandler
    })
    
    t.Run("GetUser", func(t *testing.T) {
        // Test only GetUserHandler
    })
    
    t.Run("UpdateUser", func(t *testing.T) {
        // Test only UpdateUserHandler
    })
}
```

### 3. Meaningful Test Data

```go
// ✅ Good - meaningful test data
func TestCreateUser(t *testing.T) {
    cmd := &CreateUserCommand{
        Name:     "Alice Smith",
        Email:    "alice.smith@example.com",
        Password: "SecurePassword123!",
    }
    // Test with realistic data
}

// ❌ Bad - meaningless test data
func TestCreateUser(t *testing.T) {
    cmd := &CreateUserCommand{
        Name:     "test",
        Email:    "test@test.com",
        Password: "test",
    }
}
```

### 4. Test Edge Cases

```go
func TestCreateUserEdgeCases(t *testing.T) {
    tests := []struct {
        name    string
        command *CreateUserCommand
        expect  string
    }{
        {"empty name", &CreateUserCommand{Name: "", Email: "test@example.com"}, "name is required"},
        {"invalid email", &CreateUserCommand{Name: "Test", Email: "invalid"}, "invalid email"},
        {"weak password", &CreateUserCommand{Name: "Test", Email: "test@example.com", Password: "123"}, "password too weak"},
        {"unicode name", &CreateUserCommand{Name: "José González", Email: "jose@example.com"}, ""},
        {"long email", &CreateUserCommand{Name: "Test", Email: strings.Repeat("a", 100) + "@example.com"}, "email too long"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test each edge case
        })
    }
}
```

## 🚀 Next Steps

1. **Add Monitoring**: [Monitoring & Metrics](./monitoring.md)
2. **Use Decorators**: [Decorators & Middleware](./decorators.md)
3. **Production Setup**: [Production Readiness](./production-ready.md)
4. **Performance Optimization**: [Performance Optimization](./performance.md)

---

**Ready for production monitoring? Continue with [Monitoring & Metrics](./monitoring.md)! 🚀** 