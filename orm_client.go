package tmorm

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ORMClient struct {
	client     *mongo.Client
	middleware []MiddlewareFunc
	config     *ClientConfig
	mu         sync.RWMutex
	closed     bool
}

type ClientConfig struct {
	// MongoDB连接配置
	URI                    string
	MaxPoolSize            uint64
	MinPoolSize            uint64
	ConnectTimeout         time.Duration
	MaxIdleTime            time.Duration
	HeartbeatInterval      time.Duration
	ServerSelectionTimeout time.Duration

	// 中间件配置
	EnableLogging bool
	EnableMetrics bool
	EnableTracing bool
	LogLevel      string
}

// DefaultClientConfig 默认客户端配置
func DefaultClientConfig(uri string) *ClientConfig {
	return &ClientConfig{
		URI:                    uri,
		MaxPoolSize:            100,
		MinPoolSize:            5,
		ConnectTimeout:         10 * time.Second,
		MaxIdleTime:            5 * time.Minute,
		HeartbeatInterval:      10 * time.Second,
		ServerSelectionTimeout: 30 * time.Second,
		EnableLogging:          true,
		EnableMetrics:          false,
		EnableTracing:          false,
		LogLevel:               "info",
	}
}

func NewORMClient(config *ClientConfig) (*ORMClient, error) {
	if config == nil {
		config = DefaultClientConfig("mongodb://localhost:27017")
	}

	// 创建MongoDB客户端选项
	clientOptions := options.Client().ApplyURI(config.URI)
	clientOptions.SetMaxPoolSize(config.MaxPoolSize)
	clientOptions.SetMinPoolSize(config.MinPoolSize)
	clientOptions.SetMaxConnIdleTime(config.MaxIdleTime)
	clientOptions.SetConnectTimeout(config.ConnectTimeout)
	clientOptions.SetHeartbeatInterval(config.HeartbeatInterval)
	clientOptions.SetServerSelectionTimeout(config.ServerSelectionTimeout)

	// 创建MongoDB客户端
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	ormClient := &ORMClient{
		client: client,
		config: config,
	}

	// 初始化默认中间件
	ormClient.initDefaultMiddleware()

	return ormClient, nil
}

// Database 获取数据库操作接口
func (c *ORMClient) Database(name string) *DatabaseWrapper {
	return &DatabaseWrapper{
		client:   c,
		name:     name,
		database: c.client.Database(name),
	}
}

//// Collection 直接获取集合操作接口
//func (c *ORMClient) Collection(database, collection string) *CollectionWrapper {
//	return c.Database(database).Collection(collection)
//}

// Use 添加中间件
func (c *ORMClient) Use(middleware ...MiddlewareFunc) *ORMClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.middleware = append(c.middleware, middleware...)
	return c
}

// StartSession 开始会话（利用mongo-driver原生session管理）
func (c *ORMClient) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return c.client.StartSession(opts...)
}

// Transaction 执行事务
func (c *ORMClient) Transaction(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	session, err := c.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		_, err := session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			return nil, fn(sc)
		}, opts...)
		return err
	})
}

// executeWithMiddleware 执行带中间件的操作
func (c *ORMClient) executeWithMiddleware(ctx context.Context, operation string, filter bson.M, handler func() (interface{}, error)) (interface{}, error) {
	if len(c.middleware) == 0 {
		return handler()
	}

	// 构建中间件链
	var execute func(int) (interface{}, error)
	execute = func(index int) (interface{}, error) {
		if index >= len(c.middleware) {
			return handler()
		}

		middleware := c.middleware[index]
		return middleware(ctx, operation, filter, func() (interface{}, error) {
			return execute(index + 1)
		})
	}

	return execute(0)
}

// Stats 获取客户端统计信息
func (c *ORMClient) Stats() ClientStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ClientStats{
		Connected:       !c.closed,
		MiddlewareCount: len(c.middleware),
		URI:             c.config.URI,
		MaxPoolSize:     c.config.MaxPoolSize,
		MinPoolSize:     c.config.MinPoolSize,
	}
}

// Close 关闭客户端
func (c *ORMClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.client.Disconnect(context.Background())
}

// MongoClient 获取原生MongoDB客户端
func (c *ORMClient) MongoClient() *mongo.Client {
	return c.client
}

// initDefaultMiddleware 初始化默认中间件
func (c *ORMClient) initDefaultMiddleware() {
	if c.config.EnableLogging {
		c.middleware = append(c.middleware, LoggingMiddleware(c.config.LogLevel))
	}

	if c.config.EnableMetrics {
		c.middleware = append(c.middleware, MetricsMiddleware())
	}

	if c.config.EnableTracing {
		c.middleware = append(c.middleware, TracingMiddleware())
	}
}

// DatabaseWrapper 数据库包装器
type DatabaseWrapper struct {
	client   *ORMClient
	name     string
	database *mongo.Database
}

//// Collection 获取集合包装器
//func (d *DatabaseWrapper) Collection(name string) *CollectionWrapper {
//	return &CollectionWrapper{
//		database:   d,
//		name:       name,
//		collection: d.database.Collection(name),
//		filter:     bson.M{},
//	}
//}

// Name 获取数据库名称
func (d *DatabaseWrapper) Name() string {
	return d.name
}

// Client 获取ORM客户端
func (d *DatabaseWrapper) Client() *ORMClient {
	return d.client
}

// MongoDatabase 获取原生MongoDB数据库对象
func (d *DatabaseWrapper) MongoDatabase() *mongo.Database {
	return d.database
}

// Drop 删除数据库
func (d *DatabaseWrapper) Drop(ctx context.Context) error {
	return d.database.Drop(ctx)
}

// 中间件相关类型定义
type MiddlewareFunc func(ctx context.Context, operation string, filter bson.M, next func() (interface{}, error)) (interface{}, error)

// LoggingMiddleware 日志中间件
func LoggingMiddleware(level string) MiddlewareFunc {
	return func(ctx context.Context, operation string, filter bson.M, next func() (interface{}, error)) (interface{}, error) {
		start := time.Now()
		// 这里可以使用更复杂的日志库
		println("["+level+"] Starting operation:", operation, "filter:", filter)

		result, err := next()

		duration := time.Since(start)
		if err != nil {
			println("["+level+"] Operation", operation, "failed after", duration, ":", err)
		} else {
			println("["+level+"] Operation", operation, "completed in", duration)
		}

		return result, err
	}
}

// MetricsMiddleware 指标中间件
func MetricsMiddleware() MiddlewareFunc {
	return func(ctx context.Context, operation string, filter bson.M, next func() (interface{}, error)) (interface{}, error) {
		start := time.Now()

		result, err := next()

		duration := time.Since(start)
		// 这里可以集成实际的指标收集系统
		println("[METRICS] Operation:", operation, "Duration:", duration, "Success:", err == nil)

		return result, err
	}
}

// TracingMiddleware 追踪中间件
func TracingMiddleware() MiddlewareFunc {
	return func(ctx context.Context, operation string, filter bson.M, next func() (interface{}, error)) (interface{}, error) {
		// 这里可以集成分布式追踪系统如Jaeger、Zipkin等
		println("[TRACE] Starting trace for operation:", operation)

		result, err := next()

		println("[TRACE] Completed trace for operation:", operation)
		return result, err
	}
}

// ClientStats 客户端统计信息
type ClientStats struct {
	Connected       bool   `json:"connected"`
	MiddlewareCount int    `json:"middleware_count"`
	URI             string `json:"uri"`
	MaxPoolSize     uint64 `json:"max_pool_size"`
	MinPoolSize     uint64 `json:"min_pool_size"`
}

// 错误定义
var (
	ErrClientClosed = errors.New("client is closed")
)
