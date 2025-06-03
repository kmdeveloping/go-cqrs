# Production Readiness Guide

This guide covers enterprise-grade features and deployment strategies for go-cqrs in production environments.

## 🏗️ **Production Architecture Overview**

The go-cqrs library provides a layered architecture designed for production scale:

```
┌─────────────────────────────────────────────────────────────┐
│                     Clean API Layer                         │
├─────────────────────────────────────────────────────────────┤
│  AutoRegisterHandlers() | AutoRegisterWithDependencies()   │
│  AutoRegisterProductionHandlers() | NewProductionAuto...() │
├─────────────────────────────────────────────────────────────┤
│                Production Auto-Registry                     │
├─────────────────────────────────────────────────────────────┤
│  • Health Checks    • Metrics       • Retry Logic          │
│  • Timeouts        • Configuration  • Error Recovery       │
├─────────────────────────────────────────────────────────────┤
│                     Core Auto-Registry                      │
├─────────────────────────────────────────────────────────────┤
│  • Reflection       • Type Detection • Handler Registration │
│  • DI Container     • Validation     • Thread Safety       │
├─────────────────────────────────────────────────────────────┤
│                    CQRS Manager Integration                 │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 **Production Setup**

### **Basic Production Configuration**

```go
func SetupProductionCQRS() {
    // Production configuration
    config := cqrs.DefaultProductionConfig()
    config.MaxRegistrationTime = time.Second * 10
    config.EnableRetry = true
    config.MaxRetries = 3
    config.EnableHealthChecks = true
    config.EnableMetrics = true
    
    // Create production manager
    manager := cqrs.NewCqrsManager()
    
    // Add production decorators
    manager.AddMetricsDecorator()  // Essential for monitoring
    
    // Only add logging in debug mode
    if os.Getenv("DEBUG") == "true" {
        manager.AddLoggingDecorator()
    }
    
    // Add timeout protection
    manager.AddDecorator(cqrs.TimeoutDecorator(30 * time.Second))
    
    if err := cqrs.SetManager(manager); err != nil {
        log.Fatal("Failed to set CQRS manager:", err)
    }
    
    // Production auto-registry with monitoring
    prodRegistry := cqrs.NewProductionAutoRegistry(manager, config)
    
    // Register handlers with production features
    result := prodRegistry.RegisterHandlerInstancesWithConfig(
        &CreateUserHandler{},
        &GetUserHandler{},
        &UserCreatedHandler{},
    )
    
    if len(result.Errors) > 0 {
        log.Fatalf("Failed to register handlers: %v", result.Errors)
    }
    
    log.Printf("✅ Production CQRS initialized: %d handlers, %d validators", 
        result.RegisteredHandlers, result.RegisteredValidators)
}
```

### **Environment-Specific Configuration**

```go
// config/production.go
func NewProductionConfig() *cqrs.ProductionConfig {
    return &cqrs.ProductionConfig{
        MaxRegistrationTime:    time.Second * 10,
        EnableRetry:           true,
        MaxRetries:            3,
        RetryDelay:           time.Millisecond * 100,
        EnableHealthChecks:    true,
        HealthCheckInterval:   time.Minute * 5,
        EnableMetrics:         true,
        MetricsInterval:       time.Second * 30,
        EnableCircuitBreaker:  true,
        CircuitBreakerThreshold: 5,
        TimeoutDuration:       time.Second * 30,
    }
}

// config/development.go
func NewDevelopmentConfig() *cqrs.ProductionConfig {
    return &cqrs.ProductionConfig{
        MaxRegistrationTime:    time.Second * 30, // More relaxed
        EnableRetry:           false,              // Fail fast in dev
        EnableHealthChecks:    false,              // Not needed in dev
        EnableMetrics:         true,               // Still useful
        EnableVerboseLogging:  true,              // Debug information
        TimeoutDuration:       time.Minute * 5,   // Longer for debugging
    }
}
```

## 📊 **Monitoring & Observability**

### **Built-in Metrics Collection**

```go
func SetupMonitoring() {
    // HTTP endpoint for metrics
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := cqrs.GetMetrics()
        json.NewEncoder(w).Encode(map[string]interface{}{
            "commands_executed": cqrs.GetCommandCount(),
            "queries_executed":  cqrs.GetQueryCount(),
            "events_published":  cqrs.GetEventCount(),
            "handler_counts":    getHandlerCounts(),
            "performance":       getPerformanceMetrics(),
        })
    })
    
    // Health check endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        health := cqrs.GetHealthCheck()
        if health.Status != "healthy" {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        json.NewEncoder(w).Encode(health)
    })
    
    // Readiness check
    http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
        cmdHandlers, qryHandlers, evtHandlers, validators := cqrs.GetHandlerCounts()
        ready := map[string]interface{}{
            "ready": cmdHandlers > 0 && qryHandlers > 0,
            "handlers": map[string]int{
                "commands":   cmdHandlers,
                "queries":    qryHandlers, 
                "events":     evtHandlers,
                "validators": validators,
            },
        }
        json.NewEncoder(w).Encode(ready)
    })
}

func getHandlerCounts() map[string]int {
    commands, queries, events, validators := cqrs.GetHandlerCounts()
    return map[string]int{
        "commands":   commands,
        "queries":    queries,
        "events":     events,
        "validators": validators,
    }
}
```

### **Custom Metrics Integration**

```go
// Prometheus integration
func SetupPrometheusMetrics() {
    commandCounter := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cqrs_commands_total",
            Help: "Total number of CQRS commands executed",
        },
        []string{"command_type", "status"},
    )
    
    queryCounter := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cqrs_queries_total", 
            Help: "Total number of CQRS queries executed",
        },
        []string{"query_type", "status"},
    )
    
    handlerDuration := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cqrs_handler_duration_seconds",
            Help: "Duration of CQRS handler execution",
        },
        []string{"handler_type", "message_type"},
    )
    
    prometheus.MustRegister(commandCounter, queryCounter, handlerDuration)
    
    // Custom metrics decorator
    manager.AddDecorator(PrometheusMetricsDecorator(
        commandCounter, queryCounter, handlerDuration))
}
```

### **Logging Integration**

```go
// Structured logging with context
func SetupStructuredLogging() {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(logrus.InfoLevel)
    
    if os.Getenv("ENV") == "production" {
        logger.SetLevel(logrus.WarnLevel) // Reduce noise in production
    }
    
    // Custom logging decorator
    manager.AddDecorator(StructuredLoggingDecorator(logger))
}

func StructuredLoggingDecorator(logger *logrus.Logger) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            start := time.Now()
            messageType := reflect.TypeOf(message).Name()
            
            // Extract request ID from context
            requestID, _ := ctx.Value("requestID").(string)
            
            logger.WithFields(logrus.Fields{
                "request_id":   requestID,
                "message_type": messageType,
                "timestamp":    start,
            }).Info("Handler execution started")
            
            result, err := next.Handle(ctx, message)
            
            duration := time.Since(start)
            
            logEntry := logger.WithFields(logrus.Fields{
                "request_id":   requestID,
                "message_type": messageType,
                "duration_ms":  duration.Milliseconds(),
                "success":      err == nil,
            })
            
            if err != nil {
                logEntry.WithError(err).Error("Handler execution failed")
            } else {
                logEntry.Info("Handler execution completed")
            }
            
            return result, err
        })
    }
}
```

## 🛡️ **Production Safety Features**

### **Circuit Breaker Pattern**

```go
func CircuitBreakerDecorator(threshold int, resetTimeout time.Duration) decorators.HandlerDecorator {
    breaker := NewCircuitBreaker(threshold, resetTimeout)
    
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            if breaker.IsOpen() {
                return nil, fmt.Errorf("circuit breaker is open")
            }
            
            result, err := next.Handle(ctx, message)
            
            if err != nil {
                breaker.RecordFailure()
            } else {
                breaker.RecordSuccess()
            }
            
            return result, err
        })
    }
}
```

### **Rate Limiting**

```go
func RateLimitDecorator(rate int, burst int) decorators.HandlerDecorator {
    limiter := rate.NewLimiter(rate.Limit(rate), burst)
    
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            if !limiter.Allow() {
                return nil, fmt.Errorf("rate limit exceeded")
            }
            
            return next.Handle(ctx, message)
        })
    }
}
```

### **Retry Logic with Exponential Backoff**

```go
func RetryDecorator(maxRetries int, baseDelay time.Duration) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            var lastErr error
            
            for attempt := 0; attempt <= maxRetries; attempt++ {
                if attempt > 0 {
                    // Exponential backoff
                    delay := baseDelay * time.Duration(1<<uint(attempt-1))
                    
                    select {
                    case <-ctx.Done():
                        return nil, ctx.Err()
                    case <-time.After(delay):
                    }
                }
                
                result, err := next.Handle(ctx, message)
                if err == nil {
                    return result, nil
                }
                
                lastErr = err
                
                // Check if error is retryable
                if !isRetryableError(err) {
                    break
                }
            }
            
            return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
        })
    }
}

func isRetryableError(err error) bool {
    // Define which errors are retryable
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        return true
    case errors.Is(err, syscall.ECONNRESET):
        return true
    case errors.Is(err, syscall.ECONNREFUSED):
        return true
    default:
        return false
    }
}
```

## 🔧 **Database & External Services**

### **Connection Pooling**

```go
type ProductionUserHandler struct {
    dbPool   *sql.DB
    cache    *redis.Client
    logger   *logrus.Logger
    metrics  *prometheus.CounterVec
}

func NewProductionUserHandler(dbPool *sql.DB, cache *redis.Client) *ProductionUserHandler {
    return &ProductionUserHandler{
        dbPool:  dbPool,
        cache:   cache,
        logger:  logrus.New(),
        metrics: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "user_operations_total",
        }, []string{"operation", "status"}),
    }
}

func (h *ProductionUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Use connection from pool with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Check cache first
    if cached, err := h.checkCache(ctx, cmd.Email); err == nil && cached {
        h.metrics.WithLabelValues("create_user", "cache_hit").Inc()
        return errors.New("user already exists")
    }
    
    // Begin transaction
    tx, err := h.dbPool.BeginTx(ctx, nil)
    if err != nil {
        h.metrics.WithLabelValues("create_user", "db_error").Inc()
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Create user with proper error handling
    userID, err := h.createUserInDB(ctx, tx, cmd)
    if err != nil {
        h.metrics.WithLabelValues("create_user", "create_error").Inc()
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    // Update cache
    if err := h.updateCache(ctx, cmd.Email, userID); err != nil {
        h.logger.WithError(err).Warn("Failed to update cache")
        // Don't fail the operation for cache errors
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        h.metrics.WithLabelValues("create_user", "commit_error").Inc()
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    h.metrics.WithLabelValues("create_user", "success").Inc()
    return nil
}
```

### **Graceful Shutdown**

```go
func SetupGracefulShutdown() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Shutting down gracefully...")
        
        // Create shutdown context with timeout
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        // Stop accepting new requests
        cqrs.StopAcceptingRequests()
        
        // Wait for existing requests to complete
        cqrs.WaitForPendingRequests(ctx)
        
        // Close database connections
        if dbPool != nil {
            dbPool.Close()
        }
        
        // Close cache connections
        if redisClient != nil {
            redisClient.Close()
        }
        
        log.Println("Shutdown complete")
        os.Exit(0)
    }()
}
```

## 🐳 **Container Deployment**

### **Docker Configuration**

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./main"]
```

### **Kubernetes Deployment**

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cqrs-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cqrs-app
  template:
    metadata:
      labels:
        app: cqrs-app
    spec:
      containers:
      - name: cqrs-app
        image: your-registry/cqrs-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DB_CONNECTION
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: connection-string
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: cqrs-app-service
spec:
  selector:
    app: cqrs-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
```

## 📈 **Performance Optimization**

### **Production Performance Settings**

```go
func OptimizeForProduction() {
    // Set Go runtime parameters
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Configure garbage collector
    debug.SetGCPercent(100) // Adjust based on memory profile
    
    // Configure memory pool sizes
    cqrs.SetConfig(cqrs.Config{
        HandlerPoolSize:    100,
        EventPoolSize:      1000,
        TypeCacheSize:      500,
        MetricsBufferSize:  1000,
    })
    
    // Use production-optimized decorators
    manager.AddDecorator(PooledMetricsDecorator())
    manager.AddDecorator(SampledLoggingDecorator(0.1)) // Sample 10%
}
```

### **Memory Management**

```go
// Object pooling for high-frequency operations
var eventPool = sync.Pool{
    New: func() interface{} {
        return &EventContext{}
    },
}

func ProcessEventWithPooling(ctx context.Context, event any) error {
    eventCtx := eventPool.Get().(*EventContext)
    defer eventPool.Put(eventCtx)
    
    eventCtx.Reset()
    eventCtx.SetEvent(event)
    eventCtx.SetContext(ctx)
    
    return cqrs.PublishEvent(ctx, event)
}
```

## 🔍 **Troubleshooting**

### **Common Production Issues**

**High Memory Usage:**
```go
// Monitor handler registration counts
func MonitorMemoryUsage() {
    ticker := time.NewTicker(time.Minute * 5)
    go func() {
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            log.Printf("Memory Stats: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d",
                bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
                
            commands, queries, events, validators := cqrs.GetHandlerCounts()
            log.Printf("Handler Counts: Commands=%d, Queries=%d, Events=%d, Validators=%d",
                commands, queries, events, validators)
        }
    }()
}
```

**Performance Degradation:**
```go
// Add performance monitoring
func MonitorPerformance() {
    manager.AddDecorator(PerformanceMonitorDecorator(100 * time.Millisecond))
}
```

## 📋 **Production Checklist**

### **Pre-Deployment**
- ✅ Health checks implemented
- ✅ Metrics collection enabled
- ✅ Logging configured for production
- ✅ Error handling and recovery implemented
- ✅ Database connection pooling configured
- ✅ Cache strategy implemented
- ✅ Load testing completed
- ✅ Security review completed

### **Deployment**
- ✅ Blue-green or canary deployment strategy
- ✅ Resource limits configured
- ✅ Monitoring and alerting setup
- ✅ Backup and recovery procedures
- ✅ Rollback plan ready
- ✅ Documentation updated

### **Post-Deployment**
- ✅ Monitor key metrics
- ✅ Verify health checks
- ✅ Check error rates
- ✅ Validate performance benchmarks
- ✅ Review logs for issues
- ✅ Update runbooks

---

**This production setup provides enterprise-grade reliability, monitoring, and performance for your CQRS applications! 🚀** 