package aggregator

import (
	"github.com/zzjha-cn/tm_orm/impl"
	"go.mongodb.org/mongo-driver/bson"
)

// DefaultAggregationOperation 默认聚合操作实现
type DefaultAggregationOperation struct {
	pipeline []bson.M
}

// NewAggregationOperation 创建新的聚合操作
func NewAggregationOperation() impl.IAggregationOperation {
	return &DefaultAggregationOperation{
		pipeline: make([]bson.M, 0),
	}
}

// GetPipeline 获取聚合管道
func (d *DefaultAggregationOperation) GetPipeline() []bson.M {
	return d.pipeline
}

// AddStage 添加管道阶段
func (d *DefaultAggregationOperation) AddStage(stage bson.M) {
	d.pipeline = append(d.pipeline, stage)
}

// HasStage 检查是否存在特定类型的阶段
func (d *DefaultAggregationOperation) HasStage(stageType string) bool {
	for _, stage := range d.pipeline {
		if _, exists := stage[stageType]; exists {
			return true
		}
	}
	return false
}

// GetStages 获取特定类型的所有阶段
func (d *DefaultAggregationOperation) GetStages(stageType string) []bson.M {
	var stages []bson.M
	for _, stage := range d.pipeline {
		if _, exists := stage[stageType]; exists {
			stages = append(stages, stage)
		}
	}
	return stages
}

// Clear 清空管道
func (d *DefaultAggregationOperation) Clear() {
	d.pipeline = make([]bson.M, 0)
}