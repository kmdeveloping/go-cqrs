# Configuration Guide

Complete configuration reference for go-cqrs library covering all environments and deployment scenarios.

## 🔧 **Basic Configuration**

### **Manager Configuration**

```go
// Basic setup
manager := cqrs.NewCqrsManager()

// Add decorators
manager.AddLoggingDecorator()     // Enable logging
manager.AddMetricsDecorator()     // Enable metrics collection

// Apply timeout protection
manager.AddDecorator(cqrs.TimeoutDecorator(30 * time.Second))

// Set global manager
if err := cqrs.SetManager(manager); err != nil {
    log.Fatal("Failed to set CQRS manager:", err)
}
```

### **Environment-Based Configuration**

```go
func SetupByEnvironment() {
    env := os.Getenv("ENV")
    
    switch env {
    case "production":
        setupProductionConfig()
    case "staging":
        setupStagingConfig()
    case "development":
        setupDevelopmentConfig()
    default:
        setupDefaultConfig()
    }
}

func setupProductionConfig() {
    manager := cqrs.NewCqrsManager()
    
    // Production-optimized decorators
    manager.AddMetricsDecorator()  // Essential for monitoring
    manager.AddDecorator(cqrs.TimeoutDecorator(10 * time.Second))
    manager.AddDecorator(cqrs.RetryDecorator(3, time.Second))
    
    // Only add logging if explicitly enabled
    if os.Getenv("ENABLE_DEBUG_LOGGING") == "true" {
        manager.AddLoggingDecorator()
    }
    
    cqrs.SetManager(manager)
}
```

## 🏭 **Production Configuration**

### **ProductionConfig Structure**

```go
type ProductionConfig struct {
    // Registration settings
    MaxRegistrationTime     time.Duration  // Max time for handler registration
    EnableRetry            bool           // Enable retry on failures
    MaxRetries             int            // Maximum retry attempts
    RetryDelay             time.Duration  // Delay between retries
    
    // Health monitoring
    EnableHealthChecks     bool           // Enable health check endpoints
    HealthCheckInterval    time.Duration  // Health check frequency
    
    // Metrics collection
    EnableMetrics          bool           // Enable metrics collection
    MetricsInterval        time.Duration  // Metrics collection frequency
    
    // Circuit breaker
    EnableCircuitBreaker   bool           // Enable circuit breaker pattern
    CircuitBreakerThreshold int           // Failure threshold for circuit breaker
    
    // Timeouts
    TimeoutDuration        time.Duration  // Default operation timeout
    
    // Logging
    EnableVerboseLogging   bool           // Enable detailed logging
}
```

### **Default Production Configuration**

```go
func DefaultProductionConfig() *ProductionConfig {
    return &ProductionConfig{
        MaxRegistrationTime:     time.Second * 10,
        EnableRetry:            true,
        MaxRetries:             3,
        RetryDelay:             time.Millisecond * 100,
        EnableHealthChecks:     true,
        HealthCheckInterval:    time.Minute * 5,
        EnableMetrics:          true,
        MetricsInterval:        time.Second * 30,
        EnableCircuitBreaker:   true,
        CircuitBreakerThreshold: 5,
        TimeoutDuration:        time.Second * 30,
        EnableVerboseLogging:   false,
    }
}
```

### **Custom Production Configuration**

```go
func NewCustomProductionConfig() *ProductionConfig {
    config := cqrs.DefaultProductionConfig()
    
    // Override specific settings
    config.MaxRetries = 5
    config.RetryDelay = time.Millisecond * 200
    config.TimeoutDuration = time.Second * 45
    config.CircuitBreakerThreshold = 10
    
    // Enable verbose logging for debugging
    if os.Getenv("DEBUG") == "true" {
        config.EnableVerboseLogging = true
    }
    
    return config
}
```

## ⚙️ **Environment-Specific Configurations**

### **Development Configuration**

```go
func NewDevelopmentConfig() *ProductionConfig {
    return &ProductionConfig{
        MaxRegistrationTime:     time.Second * 30,  // More relaxed timing
        EnableRetry:            false,              // Fail fast for debugging
        MaxRetries:             0,
        RetryDelay:             0,
        EnableHealthChecks:     false,              // Not needed in dev
        HealthCheckInterval:    0,
        EnableMetrics:          true,               // Still useful for dev
        MetricsInterval:        time.Second * 60,
        EnableCircuitBreaker:   false,              // Simpler debugging
        CircuitBreakerThreshold: 0,
        TimeoutDuration:        time.Minute * 5,    // Longer for debugging
        EnableVerboseLogging:   true,               // Full debugging info
    }
}

func SetupDevelopment() {
    config := NewDevelopmentConfig()
    
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator()      // Always log in development
    manager.AddMetricsDecorator()
    
    // Add development-friendly timeout
    manager.AddDecorator(cqrs.TimeoutDecorator(config.TimeoutDuration))
    
    cqrs.SetManager(manager)
    
    // Setup with development config
    prodRegistry := cqrs.NewProductionAutoRegistry(manager, config)
    
    log.Println("✅ Development CQRS configuration loaded")
}
```

### **Staging Configuration**

```go
func NewStagingConfig() *ProductionConfig {
    config := cqrs.DefaultProductionConfig()
    
    // Staging-specific overrides
    config.MaxRegistrationTime = time.Second * 15    // More relaxed than prod
    config.EnableVerboseLogging = true               // More logging for testing
    config.HealthCheckInterval = time.Minute * 2     // More frequent checks
    config.MetricsInterval = time.Second * 15        // More frequent metrics
    config.TimeoutDuration = time.Second * 45        // Longer timeouts
    
    return config
}

func SetupStaging() {
    config := NewStagingConfig()
    
    manager := cqrs.NewCqrsManager()
    manager.AddLoggingDecorator()  // Enable logging in staging
    manager.AddMetricsDecorator()
    
    // Add staging-appropriate decorators
    manager.AddDecorator(cqrs.TimeoutDecorator(config.TimeoutDuration))
    manager.AddDecorator(cqrs.RetryDecorator(config.MaxRetries, config.RetryDelay))
    
    cqrs.SetManager(manager)
    
    log.Println("✅ Staging CQRS configuration loaded")
}
```

### **Test Configuration**

```go
func NewTestConfig() *ProductionConfig {
    return &ProductionConfig{
        MaxRegistrationTime:     time.Second * 5,     // Fast tests
        EnableRetry:            false,                // Predictable behavior
        MaxRetries:             0,
        RetryDelay:             0,
        EnableHealthChecks:     false,                // Not needed in tests
        HealthCheckInterval:    0,
        EnableMetrics:          false,                // Reduce overhead
        MetricsInterval:        0,
        EnableCircuitBreaker:   false,                // Predictable behavior
        CircuitBreakerThreshold: 0,
        TimeoutDuration:        time.Second * 1,      // Fast timeouts
        EnableVerboseLogging:   false,                // Reduce test noise
    }
}

func SetupTesting() {
    config := NewTestConfig()
    
    manager := cqrs.NewCqrsManager()
    // Minimal decorators for testing
    manager.AddDecorator(cqrs.TimeoutDecorator(config.TimeoutDuration))
    
    cqrs.SetManager(manager)
}
```

## 🔧 **Auto-Registry Configuration**

### **Basic Auto-Registration**

```go
// Simple auto-registration
result := cqrs.AutoRegisterHandlers(
    &CreateUserHandler{},
    &GetUserHandler{},
    &UserCreatedHandler{},
)

fmt.Printf("Registered %d handlers\n", result.RegisteredHandlers)
```

### **Auto-Registration with Dependencies**

```go
func SetupWithDependencies() {
    // Setup dependency container
    container := cqrs.NewSimpleContainer()
    
    // Register dependencies
    cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{
        ConnectionString: os.Getenv("DATABASE_URL"),
    })
    
    cqrs.Register[EmailService](container, &SMTPEmailService{
        Host:     os.Getenv("SMTP_HOST"),
        Port:     getEnvAsInt("SMTP_PORT", 587),
        Username: os.Getenv("SMTP_USERNAME"),
        Password: os.Getenv("SMTP_PASSWORD"),
    })
    
    cqrs.Register[Logger](container, &StructuredLogger{
        Level: getLogLevel(),
    })
    
    // Auto-register with dependency injection
    result := cqrs.AutoRegisterWithDependencies(container,
        &CreateUserHandler{},     // Dependencies auto-injected
        &SendEmailHandler{},      // Dependencies auto-injected
        &GetUserHandler{},        // Dependencies auto-injected
        &CreateUserValidator{},   // Dependencies auto-injected
    )
    
    if len(result.Errors) > 0 {
        log.Fatalf("Registration failed: %v", result.Errors)
    }
    
    log.Printf("✅ Registered %d handlers, %d validators with DI", 
        result.RegisteredHandlers, result.RegisteredValidators)
}
```

### **Production Auto-Registration**

```go
func SetupProductionAutoRegistry() {
    config := cqrs.DefaultProductionConfig()
    config.MaxRetries = 5
    config.EnableHealthChecks = true
    
    manager := cqrs.NewCqrsManager()
    manager.AddMetricsDecorator()
    cqrs.SetManager(manager)
    
    // Production auto-registry with configuration
    prodRegistry := cqrs.NewProductionAutoRegistry(manager, config)
    
    result := prodRegistry.RegisterHandlerInstancesWithConfig(
        &CreateUserHandler{},
        &GetUserHandler{},
        &UserCreatedHandler{},
        &SendEmailHandler{},
        &CreateUserValidator{},
        &ValidateEmailValidator{},
    )
    
    if len(result.Errors) > 0 {
        log.Fatalf("Production registration failed: %v", result.Errors)
    }
    
    log.Printf("✅ Production registry: %d handlers, %d validators in %v",
        result.RegisteredHandlers,
        result.RegisteredValidators,
        result.Duration)
}
```

## 🎭 **Decorator Configuration**

### **Built-in Decorators**

```go
func ConfigureDecorators(env string) {
    manager := cqrs.NewCqrsManager()
    
    // Always add metrics
    manager.AddMetricsDecorator()
    
    // Environment-specific decorators
    switch env {
    case "production":
        // Minimal logging in production
        if os.Getenv("ENABLE_DEBUG_LOGGING") == "true" {
            manager.AddLoggingDecorator()
        }
        
        // Add production safety decorators
        manager.AddDecorator(cqrs.TimeoutDecorator(10 * time.Second))
        manager.AddDecorator(cqrs.RetryDecorator(3, time.Second))
        manager.AddDecorator(cqrs.CircuitBreakerDecorator(5, time.Minute))
        
    case "development":
        // Full logging in development
        manager.AddLoggingDecorator()
        
        // Longer timeouts for debugging
        manager.AddDecorator(cqrs.TimeoutDecorator(time.Minute))
        
    case "test":
        // Minimal decorators for testing
        manager.AddDecorator(cqrs.TimeoutDecorator(time.Second))
    }
    
    cqrs.SetManager(manager)
}
```

### **Custom Decorator Configuration**

```go
func SetupCustomDecorators() {
    manager := cqrs.NewCqrsManager()
    
    // Add built-in decorators
    manager.AddMetricsDecorator()
    
    // Add custom decorators
    manager.AddDecorator(AuthenticationDecorator())
    manager.AddDecorator(AuthorizationDecorator())
    manager.AddDecorator(AuditDecorator())
    manager.AddDecorator(RateLimitDecorator(100, 10)) // 100 req/sec, burst 10
    manager.AddDecorator(CacheDecorator(time.Minute * 5))
    
    // Add safety decorators last
    manager.AddDecorator(cqrs.TimeoutDecorator(30 * time.Second))
    manager.AddDecorator(cqrs.RetryDecorator(3, time.Second))
    
    cqrs.SetManager(manager)
}
```

## 📊 **Monitoring Configuration**

### **Health Check Configuration**

```go
func SetupHealthChecks() {
    config := cqrs.DefaultProductionConfig()
    config.EnableHealthChecks = true
    config.HealthCheckInterval = time.Minute * 2
    
    // HTTP endpoints
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        health := cqrs.GetHealthCheck()
        
        if health.Status != "healthy" {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(health)
    })
    
    http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
        commands, queries, events, validators := cqrs.GetHandlerCounts()
        
        ready := map[string]interface{}{
            "ready": commands > 0 && queries > 0,
            "handlers": map[string]int{
                "commands":   commands,
                "queries":    queries,
                "events":     events,
                "validators": validators,
            },
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ready)
    })
}
```

### **Metrics Configuration**

```go
func SetupMetrics() {
    config := cqrs.DefaultProductionConfig()
    config.EnableMetrics = true
    config.MetricsInterval = time.Second * 30
    
    // Prometheus metrics endpoint
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := map[string]interface{}{
            "commands_executed": cqrs.GetCommandCount(),
            "queries_executed":  cqrs.GetQueryCount(),
            "events_published":  cqrs.GetEventCount(),
            "handler_counts":    getHandlerCounts(),
            "uptime_seconds":    time.Since(startTime).Seconds(),
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(metrics)
    })
}
```

## 🔒 **Security Configuration**

### **Authentication & Authorization**

```go
func SetupSecurity() {
    manager := cqrs.NewCqrsManager()
    
    // Add security decorators
    manager.AddDecorator(AuthenticationDecorator())
    manager.AddDecorator(AuthorizationDecorator())
    manager.AddDecorator(AuditLoggingDecorator())
    
    // Rate limiting for security
    manager.AddDecorator(RateLimitDecorator(
        getEnvAsInt("RATE_LIMIT_RPS", 100),
        getEnvAsInt("RATE_LIMIT_BURST", 10),
    ))
    
    cqrs.SetManager(manager)
}

func AuthenticationDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            // Extract and validate JWT token
            token, ok := ctx.Value("Authorization").(string)
            if !ok || token == "" {
                return nil, errors.New("authentication required")
            }
            
            // Validate token and add user info to context
            userID, err := validateJWT(token)
            if err != nil {
                return nil, fmt.Errorf("invalid token: %w", err)
            }
            
            ctx = context.WithValue(ctx, "userID", userID)
            return next.Handle(ctx, message)
        })
    }
}
```

## 🧪 **Testing Configuration**

### **Unit Test Configuration**

```go
func SetupTestEnvironment() {
    // Reset any global state
    cqrs.ResetManager()
    
    // Minimal test manager
    manager := cqrs.NewCqrsManager()
    manager.AddDecorator(cqrs.TimeoutDecorator(time.Second))
    
    cqrs.SetManager(manager)
}

func TestWithCustomConfig(t *testing.T) {
    // Setup
    defer cqrs.ResetManager()
    
    config := &cqrs.ProductionConfig{
        MaxRegistrationTime:  time.Second,
        EnableRetry:         false,
        TimeoutDuration:     time.Millisecond * 500,
        EnableVerboseLogging: false,
    }
    
    manager := cqrs.NewCqrsManager()
    manager.AddDecorator(cqrs.TimeoutDecorator(config.TimeoutDuration))
    cqrs.SetManager(manager)
    
    // Test implementation
    cqrs.RegisterCommandHandler(&TestCommandHandler{})
    
    err := cqrs.ExecuteCommand(context.Background(), &TestCommand{})
    assert.NoError(t, err)
}
```

## 🔧 **Utility Functions**

### **Environment Variable Helpers**

```go
func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func getLogLevel() logrus.Level {
    switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
    case "debug":
        return logrus.DebugLevel
    case "info":
        return logrus.InfoLevel
    case "warn", "warning":
        return logrus.WarnLevel
    case "error":
        return logrus.ErrorLevel
    default:
        return logrus.InfoLevel
    }
}
```

### **Configuration Validation**

```go
func ValidateConfig(config *ProductionConfig) error {
    if config.MaxRegistrationTime <= 0 {
        return errors.New("MaxRegistrationTime must be positive")
    }
    
    if config.EnableRetry && config.MaxRetries <= 0 {
        return errors.New("MaxRetries must be positive when retry is enabled")
    }
    
    if config.EnableCircuitBreaker && config.CircuitBreakerThreshold <= 0 {
        return errors.New("CircuitBreakerThreshold must be positive when circuit breaker is enabled")
    }
    
    if config.TimeoutDuration <= 0 {
        return errors.New("TimeoutDuration must be positive")
    }
    
    return nil
}
```

## 📋 **Configuration Checklist**

### **Production Deployment Checklist**

- [ ] **Environment Variables Set**
  - [ ] `ENV=production`
  - [ ] `DATABASE_URL` configured
  - [ ] `REDIS_URL` configured (if using caching)
  - [ ] `LOG_LEVEL=warn` or `LOG_LEVEL=error`
  
- [ ] **Security Configuration**
  - [ ] Authentication decorator enabled
  - [ ] Authorization decorator enabled
  - [ ] Rate limiting configured
  - [ ] Audit logging enabled
  
- [ ] **Monitoring Configuration**
  - [ ] Health checks enabled
  - [ ] Metrics collection enabled
  - [ ] Prometheus endpoints configured
  - [ ] Alerts configured
  
- [ ] **Performance Configuration**
  - [ ] Circuit breaker enabled
  - [ ] Retry logic configured
  - [ ] Timeouts configured appropriately
  - [ ] Connection pooling configured
  
- [ ] **Reliability Configuration**
  - [ ] Graceful shutdown implemented
  - [ ] Database transaction handling
  - [ ] Error recovery strategies
  - [ ] Backup strategies in place

### **Environment Configuration Matrix**

| Setting | Development | Test | Staging | Production |
|---------|-------------|------|---------|------------|
| **Logging** | Verbose | Minimal | Verbose | Error Only |
| **Retry** | Disabled | Disabled | Enabled | Enabled |
| **Timeout** | 5 minutes | 1 second | 45 seconds | 30 seconds |
| **Health Checks** | Disabled | Disabled | Enabled | Enabled |
| **Metrics** | Enabled | Disabled | Enabled | Enabled |
| **Circuit Breaker** | Disabled | Disabled | Enabled | Enabled |

---

For implementation examples, see [Examples Guide](./examples.md) and [Integration Guide](./integrations.md). 