package repo

import (
	"context"
	tmorm "tm_orm"
	collection2 "tm_orm/collection"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository 通用仓储接口
type Repository[T any] struct {
	collection *collection2.Collection[T]
}

func NewRepository[T any](cli *tmorm.ORMClient, db string, collectionName string, msList ...tmorm.MiddlewareFunc) *Repository[T] {
	return &Repository[T]{
		collection: collection2.NewCollection[T](cli, db, collectionName, msList...),
	}
}

// Collection 获取底层集合操作器
func (r *Repository[T]) Collection() *collection2.Collection[T] {
	return r.collection
}

// FindByID 根据ID查找文档
func (r *Repository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	return r.collection.Where("_id").Eq(id).FindOne(ctx)
}

// FindAll 查找所有文档
func (r *Repository[T]) FindAll(ctx context.Context, opts ...*options.FindOptions) ([]*T, error) {
	return r.collection.Find(ctx, opts...)
}

// Create 创建新文档
func (r *Repository[T]) Create(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	return r.collection.Insert(ctx, doc)
}

// CreateMany 创建多个文档
func (r *Repository[T]) CreateMany(ctx context.Context, docs []*T) (*mongo.InsertManyResult, error) {
	return r.collection.InsertMany(ctx, docs)
}

// UpdateByID 根据ID更新文档
func (r *Repository[T]) UpdateByID(ctx context.Context, id any, updates map[string]any) (*mongo.UpdateResult, error) {
	var updateBuilder *collection2.UpdateBuilder[T]

	// 构建更新操作
	for field, value := range updates {
		if updateBuilder == nil {
			updateBuilder = r.collection.Where("_id").Eq(id).Set(field, value)
		} else {
			updateBuilder = updateBuilder.Set(field, value)
		}
	}

	if updateBuilder == nil {
		return nil, mongo.ErrNoDocuments
	}

	return updateBuilder.UpdateOne(ctx)
}

// DeleteByID 根据ID删除文档
func (r *Repository[T]) DeleteByID(ctx context.Context, id any) (*mongo.DeleteResult, error) {
	return r.collection.Where("_id").Eq(id).DeleteOne(ctx)
}

// Exists 检查文档是否存在
func (r *Repository[T]) Exists(ctx context.Context, id any) (bool, error) {
	count, err := r.collection.Where("_id").Eq(id).Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count 统计文档总数
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
	return r.collection.Count(ctx)
}

// FindPage 分页查询
func (r *Repository[T]) FindPage(ctx context.Context, page, pageSize int64, opts ...*options.FindOptions) ([]*T, int64, error) {
	// 计算总数
	total, err := r.collection.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 设置分页选项
	skip := (page - 1) * pageSize
	findOpts := options.Find().SetSkip(skip).SetLimit(pageSize)

	// 合并用户提供的选项
	if len(opts) > 0 {
		findOpts = options.MergeFindOptions(append([]*options.FindOptions{findOpts}, opts...)...)
	}

	// 执行查询
	results, err := r.collection.Find(ctx, findOpts)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// BaseModel 基础模型，包含常用字段
type BaseModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
	Deleted   bool               `bson:"deleted" json:"deleted"`
}

// GetID 获取文档ID
func (m *BaseModel) GetID() primitive.ObjectID {
	return m.ID
}

// SetID 设置文档ID
func (m *BaseModel) SetID(id primitive.ObjectID) {
	m.ID = id
}

// IsNew 判断是否为新文档
func (m *BaseModel) IsNew() bool {
	return m.ID.IsZero()
}

// BaseRepository 基础仓储，提供更多便利方法
type BaseRepository[T any] struct {
	*Repository[T]
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](cli *tmorm.ORMClient, db string, collectionName string, msList ...tmorm.MiddlewareFunc) *BaseRepository[T] {
	return &BaseRepository[T]{
		Repository: NewRepository[T](cli, db, collectionName, msList...),
	}
}

// FindByField 根据字段查找文档
func (r *BaseRepository[T]) FindByField(ctx context.Context, field string, value any) ([]*T, error) {
	return r.collection.Where(field).Eq(value).Find(ctx)
}

// FindOneByField 根据字段查找单个文档
func (r *BaseRepository[T]) FindOneByField(ctx context.Context, field string, value any) (*T, error) {
	return r.collection.Where(field).Eq(value).FindOne(ctx)
}

// UpdateByField 根据字段更新文档
func (r *BaseRepository[T]) UpdateByField(ctx context.Context, field string, fieldValue any, updates map[string]any) (*mongo.UpdateResult, error) {
	var updateBuilder *collection2.UpdateBuilder[T]

	// 构建更新操作
	for updateField, updateValue := range updates {
		if updateBuilder == nil {
			updateBuilder = r.collection.Where(field).Eq(fieldValue).Set(updateField, updateValue)
		} else {
			updateBuilder = updateBuilder.Set(updateField, updateValue)
		}
	}

	if updateBuilder == nil {
		return nil, mongo.ErrNoDocuments
	}

	return updateBuilder.Update(ctx)
}

// DeleteByField 根据字段删除文档
func (r *BaseRepository[T]) DeleteByField(ctx context.Context, field string, value any) (*mongo.DeleteResult, error) {
	return r.collection.Where(field).Eq(value).Delete(ctx)
}

// FindIn 根据字段值列表查找文档
func (r *BaseRepository[T]) FindIn(ctx context.Context, field string, values ...any) ([]*T, error) {
	return r.collection.Where(field).In(values...).Find(ctx)
}

// CountByField 根据字段统计文档数量
func (r *BaseRepository[T]) CountByField(ctx context.Context, field string, value any) (int64, error) {
	return r.collection.Where(field).Eq(value).Count(ctx)
}
