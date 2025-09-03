package tmorm

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ORMClient struct {
	client     *mongo.Client
	middleware []MiddlewareFunc
	config     *ClientConfig
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

// Use 添加中间件
func (c *ORMClient) Use(middleware ...MiddlewareFunc) *ORMClient {
	c.middleware = append(c.middleware, middleware...)
	return c
}

func (c *ORMClient) GetMiddleware() []MiddlewareFunc {
	return c.middleware
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

// Close 关闭客户端
func (c *ORMClient) Close() error {
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

// DatabaseWrapper 数据库包装器
type DatabaseWrapper struct {
	client   *ORMClient
	name     string
	database *mongo.Database
}

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
