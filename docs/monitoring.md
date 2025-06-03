# Monitoring & Metrics

Observability and health checks for production go-cqrs applications including metrics, logging, tracing, and monitoring.

## 📚 Table of Contents

- [Observability Overview](#observability-overview)
- [Metrics Collection](#metrics-collection)
- [Structured Logging](#structured-logging)
- [Distributed Tracing](#distributed-tracing)
- [Health Checks](#health-checks)
- [Performance Monitoring](#performance-monitoring)
- [Alerting](#alerting)
- [Dashboard Examples](#dashboard-examples)

## 🔍 Observability Overview

Comprehensive monitoring includes:

- **Metrics**: Quantitative measurements (latency, throughput, errors)
- **Logs**: Detailed event records with context
- **Traces**: Request flow across services
- **Health Checks**: System status and dependencies
- **Alerts**: Proactive issue notification

```go
// Enable observability in your CQRS application
func setupObservability() {
    manager := cqrs.NewCqrsManager()
    
    // Add monitoring decorators
    manager.AddLoggingDecorator()
    manager.AddMetricsDecorator()
    manager.AddTracingDecorator()
    
    cqrs.SetManager(manager)
}
```

## 📊 Metrics Collection

### Built-in Metrics

go-cqrs provides built-in metrics collection:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Command metrics
var (
    commandDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cqrs_command_duration_seconds",
            Help: "Command execution duration in seconds",
        },
        []string{"command_type", "status"},
    )
    
    commandCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cqrs_commands_total",
            Help: "Total number of commands executed",
        },
        []string{"command_type", "status"},
    )
    
    queryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cqrs_query_duration_seconds",
            Help: "Query execution duration in seconds",
        },
        []string{"query_type", "status"},
    )
    
    eventCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cqrs_events_total",
            Help: "Total number of events published",
        },
        []string{"event_type", "status"},
    )
)

func init() {
    prometheus.MustRegister(commandDuration)
    prometheus.MustRegister(commandCount)
    prometheus.MustRegister(queryDuration)
    prometheus.MustRegister(eventCount)
}
```

### Custom Metrics Decorator

```go
type MetricsDecorator struct {
    next    CommandHandler
    metrics *MetricsCollector
}

func (d *MetricsDecorator) Handle(ctx context.Context, cmd Command) error {
    start := time.Now()
    commandType := reflect.TypeOf(cmd).Name()
    
    // Execute command
    err := d.next.Handle(ctx, cmd)
    
    // Record metrics
    duration := time.Since(start)
    status := "success"
    if err != nil {
        status = "error"
    }
    
    commandDuration.WithLabelValues(commandType, status).Observe(duration.Seconds())
    commandCount.WithLabelValues(commandType, status).Inc()
    
    // Custom business metrics
    d.recordBusinessMetrics(ctx, cmd, err, duration)
    
    return err
}

func (d *MetricsDecorator) recordBusinessMetrics(ctx context.Context, cmd Command, err error, duration time.Duration) {
    switch c := cmd.(type) {
    case *CreateUserCommand:
        userRegistrations.Inc()
        if duration > 5*time.Second {
            slowUserCreations.Inc()
        }
    case *PlaceOrderCommand:
        orderValue.WithLabelValues("created").Add(float64(c.Total))
        if err != nil {
            failedOrders.Inc()
        }
    }
}
```

### Application Metrics

```go
// Business metrics
var (
    userRegistrations = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "app_user_registrations_total",
            Help: "Total number of user registrations",
        },
    )
    
    orderValue = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_order_value_total",
            Help: "Total order value",
        },
        []string{"status"},
    )
    
    activeUsers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "app_active_users",
            Help: "Number of currently active users",
        },
    )
    
    paymentLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "app_payment_processing_seconds",
            Help: "Payment processing latency",
            Buckets: prometheus.LinearBuckets(0, 0.5, 10),
        },
    )
)

// Update metrics in handlers
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    timer := prometheus.NewTimer(paymentLatency)
    defer timer.ObserveDuration()
    
    // Business logic
    err := h.processUser(ctx, cmd)
    
    if err == nil {
        userRegistrations.Inc()
        activeUsers.Inc()
    }
    
    return err
}
```

## 📝 Structured Logging

### Logger Interface

```go
type Logger interface {
    Debug(msg string, fields map[string]interface{})
    Info(msg string, fields map[string]interface{})
    Warn(msg string, fields map[string]interface{})
    Error(msg string, err error, fields map[string]interface{})
}

type StructuredLogger struct {
    logger *logrus.Logger
}

func (l *StructuredLogger) Info(msg string, fields map[string]interface{}) {
    l.logger.WithFields(logrus.Fields(fields)).Info(msg)
}

func (l *StructuredLogger) Error(msg string, err error, fields map[string]interface{}) {
    if fields == nil {
        fields = make(map[string]interface{})
    }
    fields["error"] = err.Error()
    l.logger.WithFields(logrus.Fields(fields)).Error(msg)
}
```

### Logging Decorator

```go
type LoggingDecorator struct {
    next   CommandHandler
    logger Logger
}

func (d *LoggingDecorator) Handle(ctx context.Context, cmd Command) error {
    start := time.Now()
    correlationID := getCorrelationID(ctx)
    userID := getUserID(ctx)
    commandType := reflect.TypeOf(cmd).Name()
    
    // Log command start
    d.logger.Info("Command started", map[string]interface{}{
        "correlation_id": correlationID,
        "user_id":       userID,
        "command_type":  commandType,
        "timestamp":     start,
    })
    
    // Execute command
    err := d.next.Handle(ctx, cmd)
    
    duration := time.Since(start)
    
    // Log completion
    if err != nil {
        d.logger.Error("Command failed", err, map[string]interface{}{
            "correlation_id": correlationID,
            "user_id":       userID,
            "command_type":  commandType,
            "duration_ms":   duration.Milliseconds(),
        })
    } else {
        d.logger.Info("Command completed", map[string]interface{}{
            "correlation_id": correlationID,
            "user_id":       userID,
            "command_type":  commandType,
            "duration_ms":   duration.Milliseconds(),
        })
    }
    
    return err
}
```

### Contextual Logging

```go
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    logger := getLoggerFromContext(ctx)
    
    logger.Info("Creating new user", map[string]interface{}{
        "email":           cmd.Email,
        "name":           cmd.Name,
        "registration_ip": getClientIP(ctx),
        "user_agent":     getUserAgent(ctx),
    })
    
    // Business logic
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        logger.Error("Failed to save user to database", err, map[string]interface{}{
            "email": cmd.Email,
            "name":  cmd.Name,
        })
        return err
    }
    
    logger.Info("User created successfully", map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
    })
    
    return nil
}
```

## 🔍 Distributed Tracing

### OpenTelemetry Integration

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/attribute"
)

type TracingDecorator struct {
    next   CommandHandler
    tracer trace.Tracer
}

func NewTracingDecorator(next CommandHandler) *TracingDecorator {
    return &TracingDecorator{
        next:   next,
        tracer: otel.Tracer("cqrs-commands"),
    }
}

func (d *TracingDecorator) Handle(ctx context.Context, cmd Command) error {
    commandType := reflect.TypeOf(cmd).Name()
    
    // Start span
    ctx, span := d.tracer.Start(ctx, fmt.Sprintf("command.%s", commandType))
    defer span.End()
    
    // Add attributes
    span.SetAttributes(
        attribute.String("command.type", commandType),
        attribute.String("correlation_id", getCorrelationID(ctx)),
        attribute.Int("user_id", getUserID(ctx)),
    )
    
    // Add command-specific attributes
    d.addCommandAttributes(span, cmd)
    
    // Execute command
    err := d.next.Handle(ctx, cmd)
    
    // Record outcome
    if err != nil {
        span.RecordError(err)
        span.SetAttributes(attribute.String("command.status", "error"))
    } else {
        span.SetAttributes(attribute.String("command.status", "success"))
    }
    
    return err
}

func (d *TracingDecorator) addCommandAttributes(span trace.Span, cmd Command) {
    switch c := cmd.(type) {
    case *CreateUserCommand:
        span.SetAttributes(
            attribute.String("user.email", c.Email),
            attribute.String("user.name", c.Name),
        )
    case *PlaceOrderCommand:
        span.SetAttributes(
            attribute.Int("order.customer_id", c.CustomerID),
            attribute.Int("order.item_count", len(c.Items)),
            attribute.Float64("order.total", float64(c.Total)),
        )
    }
}
```

### Trace Propagation

```go
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Child span for database operation
    ctx, span := otel.Tracer("user-service").Start(ctx, "save_user")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("db.operation", "INSERT"),
        attribute.String("db.table", "users"),
    )
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        span.RecordError(err)
        return err
    }
    
    // Child span for email service
    ctx, emailSpan := otel.Tracer("email-service").Start(ctx, "send_welcome_email")
    defer emailSpan.End()
    
    return h.emailService.SendWelcome(ctx, user.Email, user.Name)
}
```

## 🏥 Health Checks

### System Health Check

```go
type HealthChecker struct {
    userRepo      UserRepository
    emailService  EmailService
    cache         CacheService
    db           *sql.DB
}

type HealthStatus struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]Check  `json:"checks"`
}

type Check struct {
    Status  string        `json:"status"`
    Message string        `json:"message,omitempty"`
    Latency time.Duration `json:"latency"`
}

func (h *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
    checks := map[string]Check{}
    overallStatus := "healthy"
    
    // Database health check
    dbCheck := h.checkDatabase(ctx)
    checks["database"] = dbCheck
    if dbCheck.Status != "healthy" {
        overallStatus = "unhealthy"
    }
    
    // Cache health check
    cacheCheck := h.checkCache(ctx)
    checks["cache"] = cacheCheck
    if cacheCheck.Status != "healthy" {
        overallStatus = "degraded"
    }
    
    // Email service health check
    emailCheck := h.checkEmailService(ctx)
    checks["email"] = emailCheck
    if emailCheck.Status != "healthy" {
        overallStatus = "degraded"
    }
    
    return HealthStatus{
        Status:    overallStatus,
        Timestamp: time.Now(),
        Checks:    checks,
    }
}

func (h *HealthChecker) checkDatabase(ctx context.Context) Check {
    start := time.Now()
    
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    err := h.db.PingContext(ctx)
    latency := time.Since(start)
    
    if err != nil {
        return Check{
            Status:  "unhealthy",
            Message: err.Error(),
            Latency: latency,
        }
    }
    
    return Check{
        Status:  "healthy",
        Latency: latency,
    }
}

func (h *HealthChecker) checkCache(ctx context.Context) Check {
    start := time.Now()
    
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    
    testKey := "health_check_" + time.Now().Format("20060102150405")
    err := h.cache.Set(ctx, testKey, "test", time.Minute)
    if err != nil {
        return Check{
            Status:  "unhealthy",
            Message: err.Error(),
            Latency: time.Since(start),
        }
    }
    
    _, err = h.cache.Get(ctx, testKey)
    latency := time.Since(start)
    
    if err != nil {
        return Check{
            Status:  "unhealthy",
            Message: err.Error(),
            Latency: latency,
        }
    }
    
    return Check{
        Status:  "healthy",
        Latency: latency,
    }
}
```

### Health Check Endpoint

```go
func setupHealthCheckEndpoint(healthChecker *HealthChecker) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        
        health := healthChecker.CheckHealth(ctx)
        
        w.Header().Set("Content-Type", "application/json")
        
        statusCode := http.StatusOK
        if health.Status == "unhealthy" {
            statusCode = http.StatusServiceUnavailable
        } else if health.Status == "degraded" {
            statusCode = http.StatusOK // Still serving traffic
        }
        
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(health)
    })
    
    // Liveness probe (simple)
    http.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    // Readiness probe (detailed)
    http.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        
        health := healthChecker.CheckHealth(ctx)
        
        if health.Status == "unhealthy" {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Ready"))
    })
}
```

## 📈 Performance Monitoring

### Resource Monitoring

```go
import (
    "runtime"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
)

var (
    cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "app_cpu_usage_percent",
        Help: "Current CPU usage percentage",
    })
    
    memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "app_memory_usage_bytes",
        Help: "Current memory usage in bytes",
    })
    
    goroutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "app_goroutines_count",
        Help: "Current number of goroutines",
    })
)

func startResourceMonitoring() {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            updateResourceMetrics()
        }
    }()
}

func updateResourceMetrics() {
    // CPU usage
    cpuPercent, err := cpu.Percent(time.Second, false)
    if err == nil && len(cpuPercent) > 0 {
        cpuUsage.Set(cpuPercent[0])
    }
    
    // Memory usage
    memStat, err := mem.VirtualMemory()
    if err == nil {
        memoryUsage.Set(float64(memStat.Used))
    }
    
    // Goroutines
    goroutineCount.Set(float64(runtime.NumGoroutine()))
}
```

### Database Performance

```go
type DatabaseMetrics struct {
    connectionPool *sql.DB
}

var (
    dbConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "db_connections",
            Help: "Database connection pool status",
        },
        []string{"state"},
    )
    
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
            Help: "Database query duration",
        },
        []string{"operation", "table"},
    )
)

func (d *DatabaseMetrics) startMonitoring() {
    go func() {
        ticker := time.NewTicker(15 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            stats := d.connectionPool.Stats()
            
            dbConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
            dbConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
            dbConnections.WithLabelValues("idle").Set(float64(stats.Idle))
            dbConnections.WithLabelValues("wait_count").Set(float64(stats.WaitCount))
        }
    }()
}

// Wrap database operations with metrics
func (r *PostgreSQLUserRepository) Save(ctx context.Context, user *User) error {
    timer := prometheus.NewTimer(dbQueryDuration.WithLabelValues("INSERT", "users"))
    defer timer.ObserveDuration()
    
    // Database operation
    return r.saveUser(ctx, user)
}
```

## 🚨 Alerting

### Alert Rules (Prometheus)

```yaml
groups:
  - name: cqrs-application
    rules:
      - alert: HighCommandErrorRate
        expr: rate(cqrs_commands_total{status="error"}[5m]) / rate(cqrs_commands_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High command error rate"
          description: "Command error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

      - alert: SlowCommandExecution
        expr: histogram_quantile(0.95, rate(cqrs_command_duration_seconds_bucket[5m])) > 5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Slow command execution"
          description: "95th percentile command duration is {{ $value }}s"

      - alert: DatabaseConnectionPoolExhausted
        expr: db_connections{state="in_use"} / db_connections{state="open"} > 0.9
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool nearly exhausted"
          description: "{{ $value | humanizePercentage }} of database connections are in use"

      - alert: HighMemoryUsage
        expr: app_memory_usage_bytes / (1024*1024*1024) > 1.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Application memory usage is {{ $value }}GB"
```

### Alert Manager

```go
type AlertManager struct {
    slackWebhook string
    emailSMTP    EmailService
}

func (a *AlertManager) SendAlert(alert Alert) error {
    switch alert.Severity {
    case "critical":
        // Send to Slack and email
        go a.sendSlackAlert(alert)
        return a.sendEmailAlert(alert)
    case "warning":
        // Send to Slack only
        return a.sendSlackAlert(alert)
    default:
        // Log only
        log.Printf("Alert: %s - %s", alert.Summary, alert.Description)
    }
    return nil
}

func (a *AlertManager) sendSlackAlert(alert Alert) error {
    message := SlackMessage{
        Channel: "#alerts",
        Text:    fmt.Sprintf("🚨 %s: %s", alert.Severity, alert.Summary),
        Attachments: []SlackAttachment{
            {
                Color: a.getAlertColor(alert.Severity),
                Fields: []SlackField{
                    {Title: "Description", Value: alert.Description, Short: false},
                    {Title: "Time", Value: alert.Timestamp.Format(time.RFC3339), Short: true},
                    {Title: "Service", Value: "CQRS App", Short: true},
                },
            },
        },
    }
    
    return a.postToSlack(message)
}
```

## 📊 Dashboard Examples

### Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "CQRS Application Monitoring",
    "panels": [
      {
        "title": "Command Execution Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(cqrs_commands_total[5m])",
            "legendFormat": "{{ command_type }}"
          }
        ]
      },
      {
        "title": "Command Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(cqrs_commands_total{status=\"error\"}[5m]) / rate(cqrs_commands_total[5m]) * 100",
            "legendFormat": "Error Rate %"
          }
        ]
      },
      {
        "title": "Command Duration (95th percentile)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(cqrs_command_duration_seconds_bucket[5m]))",
            "legendFormat": "{{ command_type }}"
          }
        ]
      },
      {
        "title": "System Resources",
        "type": "graph",
        "targets": [
          {
            "expr": "app_cpu_usage_percent",
            "legendFormat": "CPU %"
          },
          {
            "expr": "app_memory_usage_bytes / 1024 / 1024",
            "legendFormat": "Memory MB"
          }
        ]
      }
    ]
  }
}
```

### Custom Dashboard

```go
func setupMonitoringDashboard() {
    http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
        stats := gatherApplicationStats()
        
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>CQRS App Dashboard</title>
            <meta http-equiv="refresh" content="30">
        </head>
        <body>
            <h1>CQRS Application Dashboard</h1>
            
            <h2>Command Statistics</h2>
            <p>Total Commands: {{ .TotalCommands }}</p>
            <p>Failed Commands: {{ .FailedCommands }}</p>
            <p>Success Rate: {{ .SuccessRate }}%</p>
            
            <h2>Performance</h2>
            <p>Average Command Duration: {{ .AvgDuration }}ms</p>
            <p>95th Percentile: {{ .P95Duration }}ms</p>
            
            <h2>System Health</h2>
            <p>CPU Usage: {{ .CPUUsage }}%</p>
            <p>Memory Usage: {{ .MemoryUsage }}MB</p>
            <p>Goroutines: {{ .GoroutineCount }}</p>
        </body>
        </html>
        `
        
        tmpl := template.Must(template.New("dashboard").Parse(html))
        tmpl.Execute(w, stats)
    })
}
```

## 🚀 Next Steps

1. **Add Decorators**: [Decorators & Middleware](./decorators.md)
2. **Error Handling**: [Error Handling](./error-handling.md)
3. **Configuration**: [Configuration Options](./configuration.md)
4. **Production Setup**: [Production Readiness](./production-ready.md)

---

**Ready for advanced middleware? Continue with [Decorators & Middleware](./decorators.md)! 🚀** 