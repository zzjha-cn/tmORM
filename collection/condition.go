package collection

// CollectionFieldCondition 集合字段条件构建器
type CollectionFieldCondition[T any] struct {
	collection *Collection[T]
	field      string
}

// Eq 等于条件
func (f *CollectionFieldCondition[T]) Eq(value any) *Collection[T] {
	query := NewQuery().Where(f.field).Eq(value).Build()
	return f.collection.Filter(query)
}

// Ne 不等于条件
func (f *CollectionFieldCondition[T]) Ne(value any) *Collection[T] {
	query := NewQuery().Where(f.field).Ne(value).Build()
	return f.collection.Filter(query)
}

// Gt 大于条件
func (f *CollectionFieldCondition[T]) Gt(value any) *Collection[T] {
	query := NewQuery().Where(f.field).Gt(value).Build()
	return f.collection.Filter(query)
}

// Gte 大于等于条件
func (f *CollectionFieldCondition[T]) Gte(value any) *Collection[T] {
	query := NewQuery().Where(f.field).Gte(value).Build()
	return f.collection.Filter(query)
}

// Lt 小于条件
func (f *CollectionFieldCondition[T]) Lt(value any) *Collection[T] {
	query := NewQuery().Where(f.field).Lt(value).Build()
	return f.collection.Filter(query)
}

// Lte 小于等于条件
func (f *CollectionFieldCondition[T]) Lte(value any) *Collection[T] {
	query := NewQuery().Where(f.field).Lte(value).Build()
	return f.collection.Filter(query)
}

// In 包含条件
func (f *CollectionFieldCondition[T]) In(values ...any) *Collection[T] {
	query := NewQuery().Where(f.field).In(values...).Build()
	return f.collection.Filter(query)
}

// NotIn 不包含条件
func (f *CollectionFieldCondition[T]) NotIn(values ...any) *Collection[T] {
	query := NewQuery().Where(f.field).NotIn(values...).Build()
	return f.collection.Filter(query)
}

// Regex 正则表达式条件
func (f *CollectionFieldCondition[T]) Regex(pattern string) *Collection[T] {
	query := NewQuery().Where(f.field).Regex(pattern).Build()
	return f.collection.Filter(query)
}

// Exists 字段存在条件
func (f *CollectionFieldCondition[T]) Exists(exists bool) *Collection[T] {
	query := NewQuery().Where(f.field).Exists(exists).Build()
	return f.collection.Filter(query)
}
