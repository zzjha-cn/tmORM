package tmorm

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClientManager 管理多个ORM客户端实例
type ClientManager struct {
	clients     map[string]*ORMClient
	mu          sync.RWMutex
	defaultName string // 默认客户端名称
}

// ClientManagerConfig 客户端管理器配置
type ClientManagerConfig struct {
	Clients     map[string]*ClientConfig `json:"clients"`
	DefaultName string                   `json:"default_name"`
}

// NewClientManager 创建新的客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*ORMClient),
	}
}

// NewClientManagerWithConfig 使用配置创建客户端管理器
func NewClientManagerWithConfig(config *ClientManagerConfig) (*ClientManager, error) {
	manager := NewClientManager()
	manager.defaultName = config.DefaultName

	for name, clientConfig := range config.Clients {
		client, err := NewORMClient(clientConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create client %s: %w", name, err)
		}
		manager.clients[name] = client
	}

	return manager, nil
}

// AddClient 添加客户端
func (cm *ClientManager) AddClient(name string, config *ClientConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.clients[name]; exists {
		return fmt.Errorf("client %s already exists", name)
	}

	client, err := NewORMClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client %s: %w", name, err)
	}

	cm.clients[name] = client
	return nil
}

// GetClient 获取指定名称的客户端
func (cm *ClientManager) GetClient(name string) (*ORMClient, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[name]
	if !exists {
		return nil, fmt.Errorf("client %s not found", name)
	}
	return client, nil
}

// GetDefaultClient 获取默认客户端
func (cm *ClientManager) GetDefaultClient() (*ORMClient, error) {
	if cm.defaultName == "" {
		return nil, fmt.Errorf("no default client configured")
	}
	return cm.GetClient(cm.defaultName)
}

// SetDefaultClient 设置默认客户端
func (cm *ClientManager) SetDefaultClient(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.clients[name]; !exists {
		return fmt.Errorf("client %s not found", name)
	}

	cm.defaultName = name
	return nil
}

// RemoveClient 移除客户端
func (cm *ClientManager) RemoveClient(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[name]
	if !exists {
		return fmt.Errorf("client %s not found", name)
	}

	// 关闭客户端连接
	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close client %s: %w", name, err)
	}

	delete(cm.clients, name)

	// 如果删除的是默认客户端，清空默认设置
	if cm.defaultName == name {
		cm.defaultName = ""
	}

	return nil
}

// ListClients 列出所有客户端名称
func (cm *ClientManager) ListClients() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		names = append(names, name)
	}
	return names
}

// HealthCheck 检查所有客户端的健康状态
func (cm *ClientManager) HealthCheck(ctx context.Context) map[string]error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	results := make(map[string]error)
	for name, client := range cm.clients {
		err := client.MongoClient().Ping(ctx, nil)
		results[name] = err
	}
	return results
}

// Close 关闭所有客户端
func (cm *ClientManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errors []error
	for name, client := range cm.clients {
		if err := client.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close client %s: %w", name, err))
		}
	}

	cm.clients = make(map[string]*ORMClient)
	cm.defaultName = ""

	if len(errors) > 0 {
		return fmt.Errorf("errors closing clients: %v", errors)
	}
	return nil
}

// Transaction 在指定客户端上执行事务
func (cm *ClientManager) Transaction(clientName string, ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	client, err := cm.GetClient(clientName)
	if err != nil {
		return err
	}
	return client.Transaction(ctx, fn, opts...)
}

// TransactionOnDefault 在默认客户端上执行事务
func (cm *ClientManager) TransactionOnDefault(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	client, err := cm.GetDefaultClient()
	if err != nil {
		return err
	}
	return client.Transaction(ctx, fn, opts...)
}

//// Collection 从指定客户端获取集合
//func (cm *ClientManager) Collection(clientName, database, collection string) (*CollectionWrapper, error) {
//	client, err := cm.GetClient(clientName)
//	if err != nil {
//		return nil, err
//	}
//	return client.Collection(database, collection), nil
//}
//
//// DefaultCollection 从默认客户端获取集合
//func (cm *ClientManager) DefaultCollection(database, collection string) (*CollectionWrapper, error) {
//	client, err := cm.GetDefaultClient()
//	if err != nil {
//		return nil, err
//	}
//	return client.Collection(database, collection), nil
//}

// 全局客户端管理器实例
var (
	globalManager     *ClientManager
	globalManagerOnce sync.Once
)

// GlobalManager 获取全局客户端管理器
func GlobalManager() *ClientManager {
	globalManagerOnce.Do(func() {
		globalManager = NewClientManager()
	})
	return globalManager
}

// InitGlobalManager 初始化全局客户端管理器
func InitGlobalManager(config *ClientManagerConfig) error {
	manager, err := NewClientManagerWithConfig(config)
	if err != nil {
		return err
	}
	globalManager = manager
	return nil
}

// 便捷函数，使用全局管理器

// AddGlobalClient 添加全局客户端
func AddGlobalClient(name string, config *ClientConfig) error {
	return GlobalManager().AddClient(name, config)
}

// GetGlobalClient 获取全局客户端
func GetGlobalClient(name string) (*ORMClient, error) {
	return GlobalManager().GetClient(name)
}

// GetDefaultGlobalClient 获取默认全局客户端
func GetDefaultGlobalClient() (*ORMClient, error) {
	return GlobalManager().GetDefaultClient()
}

// GlobalTransaction 在全局默认客户端上执行事务
func GlobalTransaction(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	return GlobalManager().TransactionOnDefault(ctx, fn, opts...)
}
