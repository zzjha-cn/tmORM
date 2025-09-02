package v2

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ORMClient 统一的ORM客户端，整合所有功能
type ORMClient struct {
	client          *mongo.Client
	sessionManager  *SessionManager
	middleware      []MiddlewareFunc
	config          *ClientConfig
	mu              sync.RWMutex
	closed          bool
}

// ClientConfig ORM客户端配置
type ClientConfig struct {
	// MongoDB连接配置
	URI                    string
	MaxPoolSize            uint64
	MinPoolSize            uint64
	ConnectTimeout         time.Duration
	MaxIdleTime            time.Duration
	HeartbeatInterval      time.Duration
	ServerSelectionTimeout time.Duration

	// 会话管理配置
	EnableSessionPool bool
	SessionPoolSize   int

	// 中间件配置
	EnableLogging    bool
	EnableMetrics    bool
	EnableTracing    bool
	LogLevel         string
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
		EnableSessionPool:      true,
		SessionPoolSize:        50,
		EnableLogging:          true,
		EnableMetrics:          false,
		EnableTracing:          false,
		LogLevel:               "info",
	}
}

// NewORMClient 创建新的ORM客户端
func NewORMClient(config *ClientConfig) (*ORMClient, error) {
	if config == nil {
		config = DefaultClientConfig("mongodb://localhost:27017")
	}

	// 创建MongoDB客户端
	clientOptions := options.Client().ApplyURI(config.URI)
	clientOptions.SetMaxPoolSize(config.MaxPoolSize)
	clientOptions.SetMinPoolSize(config.MinPoolSize)
	clientOptions.SetMaxConnIdleTime(config.MaxIdleTime)
	clientOptions.SetConnectTimeout(config.ConnectTimeout)
	clientOptions.SetHeartbeatInterval(config.HeartbeatInterval)
	clientOptions.SetServerSelectionTimeout(config.ServerSelectionTimeout)

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

	// 初始化会话管理器（如果启用）
	if config.EnableSessionPool {
		sessionConfig := &SessionConfig{
			MaxPoolSize:            config.MaxPoolSize,
			MinPoolSize:            config.MinPoolSize,
			ConnectTimeout:         config.ConnectTimeout,
			MaxIdleTime:            config.MaxIdleTime,
			HeartbeatInterval:      config.HeartbeatInterval,
			ServerSelectionTimeout: config.ServerSelectionTimeout,
		}
		sessionManager, err := NewSessionManager(config.URI, sessionConfig)
		if err != nil {
			return nil, err
		}
		ormClient.sessionManager = sessionManager
	}

	// 初始化默认中间件
	ormClient.initDefaultMiddleware()

	return ormClient, nil
}

// Database 获取数据库操作接口
func (c *ORMClient) Database(name string) *Database {
	return &Database{
		client:   c,
		name:     name,
		database: c.client.Database(name),
	}
}

// Collection 直接获取集合操作接口
func (c *ORMClient) Collection(database, collection string) *Collection {
	return c.Database(database).Collection(collection)
}

// Repository 创建Repository
func (c *ORMClient) Repository(database, collection string) *Repository {
	return NewRepository(c.Collection(database, collection))
}

// TypedCollection 创建类型化集合
func TypedCollection[T any](client *ORMClient, database, collection string) *TypedCollection[T] {
	return NewTypedCollection[T](client.Collection(database, collection))
}

// TypedRepository 创建类型化Repository
func TypedRepository[T any](client *ORMClient, database, collection string) *TypedRepository[T] {
	return NewTypedRepository[T](client.Collection(database, collection))
}

// Use 添加中间件
func (c *ORMClient) Use(middleware ...MiddlewareFunc) *ORMClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.middleware = append(c.middleware, middleware...)
	return c
}

// GetSession 获取会话（如果启用会话池）
func (c *ORMClient) GetSession(ctx context.Context, database, collection string) (*Session, error) {
	if c.sessionManager == nil {
		return nil, ErrSessionPoolDisabled
	}
	return c.sessionManager.GetSession(ctx, database, collection)
}

// ReleaseSession 释放会话
func (c *ORMClient) ReleaseSession(session *Session) {
	if c.sessionManager != nil {
		c.sessionManager.ReleaseSession(session)
	}
}

// WithContext 创建带上下文的客户端
func (c *ORMClient) WithContext(ctx context.Context) *ContextualClient {
	return &ContextualClient{
		client: c,
		ctx:    ctx,
	}
}

// Transaction 执行事务
func (c *ORMClient) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := c.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		return session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			return nil, fn(sc)
		})
	})
}

// Stats 获取客户端统计信息
func (c *ORMClient) Stats() ClientStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := ClientStats{
		Connected:        !c.closed,
		MiddlewareCount:  len(c.middleware),
		SessionPoolEnabled: c.sessionManager != nil,
	}

	if c.sessionManager != nil {
		sessionStats := c.sessionManager.Stats()
		stats.SessionStats = &sessionStats
	}

	return stats
}

// Close 关闭客户端
func (c *ORMClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// 关闭会话管理器
	if c.sessionManager != nil {
		if err := c.sessionManager.Close(); err != nil {
			return err
		}
	}

	// 关闭MongoDB客户端
	return c.client.Disconnect(context.Background())
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

// ClientStats 客户端统计信息
type ClientStats struct {
	Connected          bool          `json:"connected"`
	MiddlewareCount    int           `json:"middleware_count"`
	SessionPoolEnabled bool          `json:"session_pool_enabled"`
	SessionStats       *SessionStats `json:"session_stats,omitempty"`
}

// ContextualClient 带上下文的客户端
type ContextualClient struct {
	client *ORMClient
	ctx    context.Context
}

// Database 获取数据库操作接口
func (cc *ContextualClient) Database(name string) *ContextualDatabase {
	return &ContextualDatabase{
		database: cc.client.Database(name),
		ctx:      cc.ctx,
	}
}

// Collection 直接获取集合操作接口
func (cc *ContextualClient) Collection(database, collection string) *ContextualCollection {
	return cc.Database(database).Collection(collection)
}

// Transaction 执行事务
func (cc *ContextualClient) Transaction(fn func(ctx context.Context) error) error {
	return cc.client.Transaction(cc.ctx, fn)
}

// 错误定义
var (
	ErrSessionPoolDisabled = errors.New("session pool is disabled")
	ErrClientClosed        = errors.New("client is closed")
)