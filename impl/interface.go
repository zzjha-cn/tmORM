package impl

import "go.mongodb.org/mongo-driver/bson"

type (
	IBsonQuery interface {
		GetBsonD() bson.D
	}
)

// IUpdateOperation 更新操作接口
type IUpdateOperation interface {
	// GetBsonD 获取 BSON 文档
	GetBsonD() bson.D
	// AddOperation 添加操作
	AddOperation(op string, field string, value any)
	// HasOperation 检查是否存在操作
	HasOperation(op string) bool
	// GetOperation 获取特定操作的值
	GetOperation(op string) (bson.D, bool)
}

// IAggregationOperation 聚合操作接口
type IAggregationOperation interface {
	// GetPipeline 获取聚合管道
	GetPipeline() []bson.M
	// AddStage 添加管道阶段
	AddStage(stage bson.M)
	// HasStage 检查是否存在特定类型的阶段
	HasStage(stageType string) bool
	// GetStages 获取特定类型的所有阶段
	GetStages(stageType string) []bson.M
	// Clear 清空管道
	Clear()
}
