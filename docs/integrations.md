# Framework Integrations

Integration examples for go-cqrs with popular Go frameworks and external systems.

## 🌐 **HTTP Framework Integrations**

### **Gin Integration**

```go
package main

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

// Middleware
func CQRSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Add request context
        ctx := cqrs.WithRequestID(c.Request.Context(), generateRequestID())
        ctx = cqrs.WithUserID(ctx, getUserFromToken(c))
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// Handlers
func createUser(c *gin.Context) {
    var cmd CreateUserCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := cqrs.ExecuteCommand(c.Request.Context(), &cmd); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func getUser(c *gin.Context) {
    userID, _ := strconv.Atoi(c.Param("id"))
    
    user, err := cqrs.ExecuteQuery[GetUserQuery, *User](c.Request.Context(), GetUserQuery{
        UserID: userID,
    })
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func setupRoutes() *gin.Engine {
    r := gin.Default()
    r.Use(CQRSMiddleware())
    
    api := r.Group("/api/v1")
    {
        api.POST("/users", createUser)
        api.GET("/users/:id", getUser)
    }
    
    return r
}
```

### **Echo Integration**

```go
package main

import (
    "net/http"
    "strconv"
    
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func createUserEcho(c echo.Context) error {
    var cmd CreateUserCommand
    if err := c.Bind(&cmd); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    if err := cqrs.ExecuteCommand(c.Request().Context(), &cmd); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}

func setupEcho() *echo.Echo {
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    e.POST("/api/users", createUserEcho)
    e.GET("/api/users/:id", getUserEcho)
    
    return e
}
```

### **Fiber Integration**

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
)

func createUserFiber(c *fiber.Ctx) error {
    var cmd CreateUserCommand
    if err := c.BodyParser(&cmd); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    
    if err := cqrs.ExecuteCommand(c.Context(), &cmd); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.Status(201).JSON(fiber.Map{"message": "User created"})
}

func setupFiber() *fiber.App {
    app := fiber.New()
    app.Use(logger.New())
    
    api := app.Group("/api/v1")
    api.Post("/users", createUserFiber)
    api.Get("/users/:id", getUserFiber)
    
    return app
}
```

## 🗄️ **Database Integrations**

### **GORM Integration**

```go
type GORMUserRepository struct {
    db *gorm.DB
}

func (r *GORMUserRepository) Create(ctx context.Context, user *User) (int, error) {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        return 0, err
    }
    return user.ID, nil
}

func (r *GORMUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

// Setup
func setupGORM() {
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, &GORMUserRepository{db: db})
    
    cqrs.AutoRegisterWithDependencies(container, &CreateUserHandler{})
}
```

### **MongoDB Integration**

```go
type MongoUserRepository struct {
    collection *mongo.Collection
}

func (r *MongoUserRepository) Create(ctx context.Context, user *User) (string, error) {
    result, err := r.collection.InsertOne(ctx, user)
    if err != nil {
        return "", err
    }
    return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    objID, _ := primitive.ObjectIDFromHex(id)
    var user User
    err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
    return &user, err
}
```

### **Redis Integration**

```go
type RedisCacheService struct {
    client *redis.Client
}

func (s *RedisCacheService) Get(ctx context.Context, key string) ([]byte, error) {
    return s.client.Get(ctx, key).Bytes()
}

func (s *RedisCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return s.client.Set(ctx, key, value, ttl).Err()
}

// Cached query handler
type CachedGetUserHandler struct {
    UserRepo UserRepository `inject:""`
    Cache    CacheService   `inject:""`
}

func (h *CachedGetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", q.UserID)
    
    // Try cache first
    if cached, err := h.Cache.Get(ctx, cacheKey); err == nil {
        var user User
        json.Unmarshal(cached, &user)
        return &user, nil
    }
    
    // Get from database
    user, err := h.UserRepo.GetByID(ctx, q.UserID)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    if data, err := json.Marshal(user); err == nil {
        h.Cache.Set(ctx, cacheKey, data, time.Minute*15)
    }
    
    return user, nil
}
```

## 📨 **Message Queue Integrations**

### **RabbitMQ Integration**

```go
type RabbitMQEventPublisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (p *RabbitMQEventPublisher) PublishEvent(ctx context.Context, event IEvent) error {
    body, _ := json.Marshal(event)
    
    return p.channel.Publish(
        "events",                    // exchange
        reflect.TypeOf(event).Name(), // routing key
        false,                       // mandatory
        false,                       // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
}

// Event handler that publishes to queue
type QueueEventHandler struct {
    Publisher EventPublisher `inject:""`
}

func (h *QueueEventHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    return h.Publisher.PublishEvent(ctx, e)
}
```

### **Apache Kafka Integration**

```go
type KafkaEventPublisher struct {
    producer *kafka.Writer
}

func (p *KafkaEventPublisher) PublishEvent(ctx context.Context, event IEvent) error {
    message, _ := json.Marshal(event)
    
    return p.producer.WriteMessages(ctx, kafka.Message{
        Topic: "domain-events",
        Key:   []byte(reflect.TypeOf(event).Name()),
        Value: message,
    })
}

func setupKafka() {
    producer := &kafka.Writer{
        Addr:     kafka.TCP("localhost:9092"),
        Topic:    "domain-events",
        Balancer: &kafka.LeastBytes{},
    }
    
    container := cqrs.NewSimpleContainer()
    cqrs.Register[EventPublisher](container, &KafkaEventPublisher{producer: producer})
}
```

## ☁️ **Cloud Service Integrations**

### **AWS S3 Integration**

```go
type S3FileService struct {
    client *s3.Client
    bucket string
}

func (s *S3FileService) UploadFile(ctx context.Context, key string, data []byte) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
        Body:   bytes.NewReader(data),
    })
    return err
}

type UploadFileHandler struct {
    FileService S3FileService `inject:""`
}

func (h *UploadFileHandler) Handle(ctx context.Context, cmd *UploadFileCommand) error {
    return h.FileService.UploadFile(ctx, cmd.FileName, cmd.Data)
}
```

### **AWS SQS Integration**

```go
type SQSEventPublisher struct {
    client   *sqs.Client
    queueURL string
}

func (p *SQSEventPublisher) PublishEvent(ctx context.Context, event IEvent) error {
    body, _ := json.Marshal(event)
    
    _, err := p.client.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    aws.String(p.queueURL),
        MessageBody: aws.String(string(body)),
    })
    return err
}
```

### **Google Pub/Sub Integration**

```go
type PubSubEventPublisher struct {
    topic *pubsub.Topic
}

func (p *PubSubEventPublisher) PublishEvent(ctx context.Context, event IEvent) error {
    data, _ := json.Marshal(event)
    
    result := p.topic.Publish(ctx, &pubsub.Message{Data: data})
    _, err := result.Get(ctx)
    return err
}
```

## 🔍 **Monitoring Integrations**

### **Prometheus Integration**

```go
var (
    commandsExecuted = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "cqrs_commands_total"},
        []string{"command_type", "status"},
    )
    
    queryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "cqrs_query_duration_seconds"},
        []string{"query_type"},
    )
)

func PrometheusDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            start := time.Now()
            messageType := reflect.TypeOf(message).Name()
            
            result, err := next.Handle(ctx, message)
            
            duration := time.Since(start)
            status := "success"
            if err != nil {
                status = "error"
            }
            
            if isCommand(message) {
                commandsExecuted.WithLabelValues(messageType, status).Inc()
            } else if isQuery(message) {
                queryDuration.WithLabelValues(messageType).Observe(duration.Seconds())
            }
            
            return result, err
        })
    }
}
```

### **OpenTelemetry Integration**

```go
func OpenTelemetryDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            messageType := reflect.TypeOf(message).Name()
            
            ctx, span := otel.Tracer("cqrs").Start(ctx, fmt.Sprintf("cqrs.%s", messageType))
            defer span.End()
            
            span.SetAttributes(
                attribute.String("cqrs.message_type", messageType),
                attribute.String("cqrs.handler_type", getHandlerType(message)),
            )
            
            result, err := next.Handle(ctx, message)
            
            if err != nil {
                span.RecordError(err)
                span.SetStatus(codes.Error, err.Error())
            }
            
            return result, err
        })
    }
}
```

## 🧪 **Testing Framework Integrations**

### **Testify Integration**

```go
package tests

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) (int, error) {
    args := m.Called(ctx, user)
    return args.Int(0), args.Error(1)
}

type UserTestSuite struct {
    suite.Suite
    mockRepo *MockUserRepository
    handler  *CreateUserHandler
}

func (s *UserTestSuite) SetupTest() {
    cqrs.ResetManager()
    s.mockRepo = &MockUserRepository{}
    s.handler = &CreateUserHandler{UserRepo: s.mockRepo}
    
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    cqrs.RegisterCommandHandler(s.handler)
}

func (s *UserTestSuite) TestCreateUser() {
    s.mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *User) bool {
        return u.Email == "test@example.com"
    })).Return(1, nil)
    
    cmd := &CreateUserCommand{
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    err := cqrs.ExecuteCommand(context.Background(), cmd)
    
    assert.NoError(s.T(), err)
    s.mockRepo.AssertExpectations(s.T())
}

func TestUserSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
```

### **Ginkgo/Gomega Integration**

```go
package tests

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("CQRS Integration", func() {
    BeforeEach(func() {
        cqrs.ResetManager()
        manager := cqrs.NewCqrsManager()
        cqrs.SetManager(manager)
    })
    
    Context("When executing commands", func() {
        It("should handle create user command", func() {
            cqrs.RegisterCommandHandler(&CreateUserHandler{})
            
            cmd := &CreateUserCommand{
                Email: "test@example.com",
                Name:  "Test User",
            }
            
            err := cqrs.ExecuteCommand(context.Background(), cmd)
            Expect(err).ToNot(HaveOccurred())
        })
    })
})
```

## 🔐 **Security Integrations**

### **JWT Integration**

```go
func JWTAuthDecorator(secretKey string) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            tokenString, ok := ctx.Value("Authorization").(string)
            if !ok {
                return nil, errors.New("no authorization token")
            }
            
            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                return []byte(secretKey), nil
            })
            
            if err != nil || !token.Valid {
                return nil, errors.New("invalid token")
            }
            
            if claims, ok := token.Claims.(jwt.MapClaims); ok {
                ctx = context.WithValue(ctx, "userID", claims["user_id"])
                ctx = context.WithValue(ctx, "userRole", claims["role"])
            }
            
            return next.Handle(ctx, message)
        })
    }
}
```

### **OAuth2 Integration**

```go
func OAuth2Decorator(provider OAuth2Provider) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            token, ok := ctx.Value("oauth_token").(string)
            if !ok {
                return nil, errors.New("no OAuth token")
            }
            
            userInfo, err := provider.ValidateToken(ctx, token)
            if err != nil {
                return nil, fmt.Errorf("invalid OAuth token: %w", err)
            }
            
            ctx = context.WithValue(ctx, "userID", userInfo.ID)
            ctx = context.WithValue(ctx, "userEmail", userInfo.Email)
            
            return next.Handle(ctx, message)
        })
    }
}
```

## 📊 **Complete Integration Example**

```go
package main

func main() {
    // Database setup
    db := setupDatabase()
    
    // Cache setup
    redisClient := setupRedis()
    
    // Message queue setup
    kafkaProducer := setupKafka()
    
    // Dependency container
    container := cqrs.NewSimpleContainer()
    cqrs.Register[UserRepository](container, &GORMUserRepository{db: db})
    cqrs.Register[CacheService](container, &RedisCacheService{client: redisClient})
    cqrs.Register[EventPublisher](container, &KafkaEventPublisher{producer: kafkaProducer})
    
    // CQRS setup with decorators
    manager := cqrs.NewCqrsManager()
    manager.AddDecorator(JWTAuthDecorator(os.Getenv("JWT_SECRET")))
    manager.AddDecorator(PrometheusDecorator())
    manager.AddDecorator(OpenTelemetryDecorator())
    manager.AddMetricsDecorator()
    
    cqrs.SetManager(manager)
    
    // Auto-register handlers
    cqrs.AutoRegisterWithDependencies(container,
        &CreateUserHandler{},
        &GetUserHandler{},
        &UserCreatedEventHandler{},
    )
    
    // HTTP server
    router := setupGinRouter()
    router.Run(":8080")
}
```

For detailed implementation examples, see [Real-World Examples](./examples.md) and [API Reference](./api-reference.md). 