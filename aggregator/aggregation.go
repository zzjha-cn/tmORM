package aggregator

import (
	"context"
	tmorm "tm_orm"
	"tm_orm/expression"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Aggregator 聚合器
type Aggregator[T any] struct {
	client         *tmorm.ORMClient
	DBName         string
	CollectionName string
	pipeline       []bson.M
	msList         []tmorm.MiddlewareFunc
}

// NewAggregator 创建新的聚合器
func NewAggregator[T any](cli *tmorm.ORMClient, db string, collectionName string, msList ...tmorm.MiddlewareFunc) *Aggregator[T] {
	return &Aggregator[T]{
		client:         cli,
		DBName:         db,
		CollectionName: collectionName,
		pipeline:       make([]bson.M, 0),
		msList:         msList,
	}
}

// Pipeline 获取管道数据
func (a *Aggregator[T]) Pipeline() []bson.M {
	return a.pipeline
}

func (a *Aggregator[T]) combineChain(handle tmorm.MiddlewareFunc) (res []tmorm.MiddlewareFunc) {
	res = append(res, a.client.GetMiddleware()...)
	res = append(res, a.msList...)
	res = append(res, handle)
	return
}

// AddStage 添加管道阶段
func (a *Aggregator[T]) AddStage(stage bson.M) *Aggregator[T] {
	a.pipeline = append(a.pipeline, stage)
	return a
}

// Match 匹配阶段
func (a *Aggregator[T]) Match(filter bson.M) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.MatchOp: filter})
}

// MatchByExpression 使用表达式匹配
func (a *Aggregator[T]) MatchByExpression(expr *expression.Expression) *Aggregator[T] {
	return a.Match(expr.Build())
}

// Group 分组阶段
func (a *Aggregator[T]) Group(id any, fields bson.M) *Aggregator[T] {
	groupStage := bson.M{"_id": id}
	for k, v := range fields {
		groupStage[k] = v
	}
	return a.AddStage(bson.M{tmorm.GroupOp: groupStage})
}

// Sort 排序阶段
func (a *Aggregator[T]) Sort(sorts bson.D) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.SortOp: sorts})
}

// SortBy 按字段排序
func (a *Aggregator[T]) SortBy(field string, order int) *Aggregator[T] {
	return a.Sort(bson.D{{field, order}})
}

// SortAsc 升序排序
func (a *Aggregator[T]) SortAsc(fields ...string) *Aggregator[T] {
	sorts := make(bson.D, 0, len(fields))
	for _, field := range fields {
		sorts = append(sorts, bson.E{Key: field, Value: 1})
	}
	return a.Sort(sorts)
}

// SortDesc 降序排序
func (a *Aggregator[T]) SortDesc(fields ...string) *Aggregator[T] {
	sorts := make(bson.D, 0, len(fields))
	for _, field := range fields {
		sorts = append(sorts, bson.E{Key: field, Value: -1})
	}
	return a.Sort(sorts)
}

// Project 投影阶段
func (a *Aggregator[T]) Project(projection bson.M) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.ProjectOp: projection})
}

// ProjectFields 投影指定字段
func (a *Aggregator[T]) ProjectFields(includeID bool, fields ...string) *Aggregator[T] {
	projection := make(bson.M)
	if !includeID {
		projection["_id"] = 0
	}
	for _, field := range fields {
		projection[field] = 1
	}
	return a.Project(projection)
}

// Limit 限制结果数量
func (a *Aggregator[T]) Limit(limit int64) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.LimitOp: limit})
}

// Skip 跳过指定数量
func (a *Aggregator[T]) Skip(skip int64) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.SkipOp: skip})
}

// Unwind 展开数组
func (a *Aggregator[T]) Unwind(path string, preserveNullAndEmptyArrays ...bool) *Aggregator[T] {
	unwindStage := bson.M{"path": "$" + path}
	if len(preserveNullAndEmptyArrays) > 0 && preserveNullAndEmptyArrays[0] {
		unwindStage["preserveNullAndEmptyArrays"] = true
	}
	return a.AddStage(bson.M{tmorm.UnwindOp: unwindStage})
}

// Lookup 关联查询
func (a *Aggregator[T]) Lookup(from, localField, foreignField, as string) *Aggregator[T] {
	return a.AddStage(bson.M{
		tmorm.StageLookUpOp: bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
}

// AddFields 添加字段
func (a *Aggregator[T]) AddFields(fields bson.M) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.AddFieldsOp: fields})
}

// ReplaceRoot 替换根文档
func (a *Aggregator[T]) ReplaceRoot(newRoot any) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.ReplaceWithOp: bson.M{"newRoot": newRoot}})
}

// Facet 多面搜索
func (a *Aggregator[T]) Facet(facets bson.M) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.FacetOp: facets})
}

// Bucket 分桶
func (a *Aggregator[T]) Bucket(groupBy any, boundaries []any, defaultBucket any, output bson.M) *Aggregator[T] {
	bucketStage := bson.M{
		"groupBy":    groupBy,
		"boundaries": boundaries,
	}
	if defaultBucket != nil {
		bucketStage["default"] = defaultBucket
	}
	if output != nil {
		bucketStage["output"] = output
	}
	return a.AddStage(bson.M{tmorm.BucketOp: bucketStage})
}

// Execute 执行聚合
func (a *Aggregator[T]) Execute(ctx context.Context, opts ...*options.AggregateOptions) ([]*T, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		// 转换为bson.A格式
		pipelineA := make(bson.A, len(a.pipeline))
		for i, stage := range a.pipeline {
			pipelineA[i] = stage
		}

		// 执行聚合
		cursor, err := a.client.Database(a.DBName).MongoDatabase().Collection(a.CollectionName).Aggregate(ctx, pipelineA, opts...)
		if err != nil {
			mctx.Result = &tmorm.MResult{Val: nil, Err: err}
			next(mctx)
			return
		}
		defer cursor.Close(ctx)

		// 解析结果
		var results []*T
		err = cursor.All(ctx, &results)
		mctx.Result = &tmorm.MResult{Val: results, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, "Aggregate")
	tmorm.Executor(mctx, a.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.([]*T), res.Err
	}
	return nil, res.Err
}

// ExecuteRaw 执行聚合并返回原始结果
func (a *Aggregator[T]) ExecuteRaw(ctx context.Context, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		pipelineA := make(bson.A, len(a.pipeline))
		for i, stage := range a.pipeline {
			pipelineA[i] = stage
		}

		cursor, err := a.client.Database(a.DBName).MongoDatabase().Collection(a.CollectionName).Aggregate(ctx, pipelineA, opts...)
		mctx.Result = &tmorm.MResult{Val: cursor, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, "AggregateRaw")
	tmorm.Executor(mctx, a.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.Cursor), res.Err
	}
	return nil, res.Err
}

// ExecuteOne 执行聚合并返回单个结果
func (a *Aggregator[T]) ExecuteOne(ctx context.Context, opts ...*options.AggregateOptions) (*T, error) {
	results, err := a.Limit(1).Execute(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return results[0], nil
}

// Count 统计聚合结果数量
func (a *Aggregator[T]) Count(field string) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.CountOp: field})
}

// GroupBuilder 分组构建器
type GroupBuilder struct {
	aggregator *Aggregator[any]
	groupID    any
	fields     bson.M
}

// NewGroupBuilder 创建分组构建器
func (a *Aggregator[T]) NewGroupBuilder(id any) *GroupBuilder {
	return &GroupBuilder{
		aggregator: (*Aggregator[any])(a),
		groupID:    id,
		fields:     make(bson.M),
	}
}

// Sum 求和
func (gb *GroupBuilder) Sum(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.SumOp: "$" + sourceField}
	return gb
}

// SumValue 求和（使用固定值）
func (gb *GroupBuilder) SumValue(field string, value any) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.SumOp: value}
	return gb
}

// Avg 平均值
func (gb *GroupBuilder) Avg(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.AvgOp: "$" + sourceField}
	return gb
}

// Count 计数
func (gb *GroupBuilder) Count(field string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.SumOp: 1}
	return gb
}

// Max 最大值
func (gb *GroupBuilder) Max(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.MaxOp: "$" + sourceField}
	return gb
}

// Min 最小值
func (gb *GroupBuilder) Min(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.MinOp: "$" + sourceField}
	return gb
}

// First 第一个值
func (gb *GroupBuilder) First(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.FirstOp: "$" + sourceField}
	return gb
}

// Last 最后一个值
func (gb *GroupBuilder) Last(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.LastOp: "$" + sourceField}
	return gb
}

// Push 推入数组
func (gb *GroupBuilder) Push(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.PushOp: "$" + sourceField}
	return gb
}

// AddToSet 添加到集合（去重）
func (gb *GroupBuilder) AddToSet(field, sourceField string) *GroupBuilder {
	gb.fields[field] = bson.M{tmorm.AddToSetOp: "$" + sourceField}
	return gb
}

// Build 构建分组阶段
func (gb *GroupBuilder) Build() *Aggregator[any] {
	return gb.aggregator.Group(gb.groupID, gb.fields)
}

// 便利方法

// GroupBy 按字段分组
func (a *Aggregator[T]) GroupBy(field string) *GroupBuilder {
	return a.NewGroupBuilder("$" + field)
}

// GroupByFields 按多个字段分组
func (a *Aggregator[T]) GroupByFields(fields ...string) *GroupBuilder {
	groupID := make(bson.M)
	for _, field := range fields {
		groupID[field] = "$" + field
	}
	return a.NewGroupBuilder(groupID)
}

// GroupAll 全局分组
func (a *Aggregator[T]) GroupAll() *GroupBuilder {
	return a.NewGroupBuilder(nil)
}

// 时间相关的分组方法

// GroupByYear 按年分组
func (a *Aggregator[T]) GroupByYear(dateField string) *GroupBuilder {
	return a.NewGroupBuilder(bson.M{"year": bson.M{tmorm.YearOp: "$" + dateField}})
}

// GroupByMonth 按月分组
func (a *Aggregator[T]) GroupByMonth(dateField string) *GroupBuilder {
	return a.NewGroupBuilder(bson.M{
		"year":  bson.M{tmorm.YearOp: "$" + dateField},
		"month": bson.M{tmorm.MonthOp: "$" + dateField},
	})
}

// GroupByDay 按日分组
func (a *Aggregator[T]) GroupByDay(dateField string) *GroupBuilder {
	return a.NewGroupBuilder(bson.M{
		"year":  bson.M{tmorm.YearOp: "$" + dateField},
		"month": bson.M{tmorm.MonthOp: "$" + dateField},
		"day":   bson.M{tmorm.DayOfMonthOp: "$" + dateField},
	})
}

// GroupByHour 按小时分组
func (a *Aggregator[T]) GroupByHour(dateField string) *GroupBuilder {
	return a.NewGroupBuilder(bson.M{
		"year":  bson.M{tmorm.YearOp: "$" + dateField},
		"month": bson.M{tmorm.MonthOp: "$" + dateField},
		"day":   bson.M{tmorm.DayOfMonthOp: "$" + dateField},
		"hour":  bson.M{tmorm.HourOp: "$" + dateField},
	})
}

// 分页聚合

// Paginate 分页
func (a *Aggregator[T]) Paginate(page, pageSize int64) *Aggregator[T] {
	skip := (page - 1) * pageSize
	return a.Skip(skip).Limit(pageSize)
}

// PaginateWithCount 分页并统计总数
func (a *Aggregator[T]) PaginateWithCount(ctx context.Context, page, pageSize int64) (*PaginationResult[T], error) {
	// 创建两个管道：一个用于获取数据，一个用于统计总数
	dataPipeline := make([]bson.M, len(a.pipeline))
	copy(dataPipeline, a.pipeline)

	countPipeline := make([]bson.M, len(a.pipeline))
	copy(countPipeline, a.pipeline)

	// 使用facet同时执行两个管道
	facetAgg := NewAggregator[any](a.client, a.DBName, a.CollectionName, a.msList...)
	for _, stage := range a.pipeline {
		facetAgg.AddStage(stage)
	}

	skip := (page - 1) * pageSize
	facetAgg.AddStage(bson.M{
		tmorm.FacetOp: bson.M{
			"data": []bson.M{
				{tmorm.SkipOp: skip},
				{tmorm.LimitOp: pageSize},
			},
			"count": []bson.M{
				{tmorm.CountOp: "total"},
			},
		},
	})

	results, err := facetAgg.Execute(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &PaginationResult[T]{
			Data:       []*T{},
			Total:      0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}, nil
	}

	// 解析结果（这里需要根据实际情况调整）
	return &PaginationResult[T]{
		Data:       []*T{}, // 需要从facet结果中解析
		Total:      0,      // 需要从facet结果中解析
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0, // 计算得出
	}, nil
}

// PaginationResult 分页结果
type PaginationResult[T any] struct {
	Data       []*T  `json:"data"`
	Total      int64 `json:"total"`
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
}

// 聚合表达式构建器

// AggExpr 聚合表达式
type AggExpr struct {
	expr any
}

// Field 字段引用
func AggField(name string) *AggExpr {
	return &AggExpr{expr: "$" + name}
}

// Literal 字面值
func AggLiteral(value any) *AggExpr {
	return &AggExpr{expr: value}
}

// Add 加法
func (e *AggExpr) Add(other *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.AddOp: []any{e.expr, other.expr}}}
}

// Subtract 减法
func (e *AggExpr) Subtract(other *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.SubtractOp: []any{e.expr, other.expr}}}
}

// Multiply 乘法
func (e *AggExpr) Multiply(other *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.MultiplyOp: []any{e.expr, other.expr}}}
}

// Divide 除法
func (e *AggExpr) Divide(other *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.DivideOp: []any{e.expr, other.expr}}}
}

// Mod 取模
func (e *AggExpr) Mod(other *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ModOp: []any{e.expr, other.expr}}}
}

// Cond 条件表达式
func AggCond(condition, ifTrue, ifFalse *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.CondOp: []any{condition.expr, ifTrue.expr, ifFalse.expr}}}
}

// IfNull 空值处理
func (e *AggExpr) IfNull(replacement *AggExpr) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.IfNullOp: []any{e.expr, replacement.expr}}}
}

// Size 数组大小
func (e *AggExpr) Size() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.SizeOp: e.expr}}
}

// Type 类型检查
func (e *AggExpr) Type() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.TypeOp: e.expr}}
}

// ToString 转换为字符串
func (e *AggExpr) ToString() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToStringOp: e.expr}}
}

// ToInt 转换为整数
func (e *AggExpr) ToInt() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToIntOp: e.expr}}
}

// ToDouble 转换为浮点数
func (e *AggExpr) ToDouble() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToDoubleOp: e.expr}}
}

// ToDate 转换为日期
func (e *AggExpr) ToDate() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToDateOp: e.expr}}
}

// GetExpr 获取表达式
func (e *AggExpr) GetExpr() any {
	return e.expr
}

// 日期表达式

// Year 年份
func (e *AggExpr) Year() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.YearOp: e.expr}}
}

// Month 月份
func (e *AggExpr) Month() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.MonthOp: e.expr}}
}

// Day 日期
func (e *AggExpr) Day() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.DayOfMonthOp: e.expr}}
}

// Hour 小时
func (e *AggExpr) Hour() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.HourOp: e.expr}}
}

// Minute 分钟
func (e *AggExpr) Minute() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.MinuteOp: e.expr}}
}

// Second 秒
func (e *AggExpr) Second() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.SecondOp: e.expr}}
}

// DateToString 日期转字符串
func (e *AggExpr) DateToString(format string, timezone ...string) *AggExpr {
	dateToString := bson.M{
		"format": format,
		"date":   e.expr,
	}
	if len(timezone) > 0 {
		dateToString["timezone"] = timezone[0]
	}
	return &AggExpr{expr: bson.M{tmorm.DateToStringOp: dateToString}}
}

// 字符串表达式

// Concat 字符串连接
func AggConcat(exprs ...*AggExpr) *AggExpr {
	args := make([]any, len(exprs))
	for i, expr := range exprs {
		args[i] = expr.expr
	}
	return &AggExpr{expr: bson.M{tmorm.ConcatOp: args}}
}

// Substr 子字符串
func (e *AggExpr) Substr(start, length int) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.SubstrBytesOp: []any{e.expr, start, length}}}
}

// ToLower 转小写
func (e *AggExpr) ToLower() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToLowerOp: e.expr}}
}

// ToUpper 转大写
func (e *AggExpr) ToUpper() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToUpperOp: e.expr}}
}

// StrLen 字符串长度
func (e *AggExpr) StrLen() *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.StrLenCPOp: e.expr}}
}
