package aggregator

import (
	"time"

	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
)

// MatchOperators 匹配操作符
type MatchOperators struct {
	aggregator *Aggregator[any]
}

// NewMatchOperators 创建匹配操作符
func (a *Aggregator[T]) NewMatchOperators() *MatchOperators {
	return &MatchOperators{
		aggregator: (*Aggregator[any])(a),
	}
}

// Eq 等于
func (m *MatchOperators) Eq(field string, value any) *MatchOperators {
	m.aggregator.Match(bson.M{field: value})
	return m
}

// Ne 不等于
func (m *MatchOperators) Ne(field string, value any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.NeOp: value}})
	return m
}

// Gt 大于
func (m *MatchOperators) Gt(field string, value any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.GtOp: value}})
	return m
}

// Gte 大于等于
func (m *MatchOperators) Gte(field string, value any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.GteOp: value}})
	return m
}

// Lt 小于
func (m *MatchOperators) Lt(field string, value any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.LtOp: value}})
	return m
}

// Lte 小于等于
func (m *MatchOperators) Lte(field string, value any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.LteOp: value}})
	return m
}

// In 在数组中
func (m *MatchOperators) In(field string, values []any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.InOp: values}})
	return m
}

// NotIn 不在数组中
func (m *MatchOperators) NotIn(field string, values []any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.NinOp: values}})
	return m
}

// Exists 字段存在
func (m *MatchOperators) Exists(field string, exists bool) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.ExistsOp: exists}})
	return m
}

// Regex 正则匹配
func (m *MatchOperators) Regex(field, pattern string, options ...string) *MatchOperators {
	regexDoc := bson.M{tmorm.RegexOp: pattern}
	if len(options) > 0 {
		regexDoc[tmorm.OptionsOp] = options[0]
	}
	m.aggregator.Match(bson.M{field: regexDoc})
	return m
}

// Size 数组大小
func (m *MatchOperators) Size(field string, size int) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.SizeOp: size}})
	return m
}

// Type 字段类型
func (m *MatchOperators) Type(field string, bsonType int) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.TypeOp: bsonType}})
	return m
}

// All 数组包含所有元素
func (m *MatchOperators) All(field string, values []any) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.AllOp: values}})
	return m
}

// ElemMatch 数组元素匹配
func (m *MatchOperators) ElemMatch(field string, condition bson.M) *MatchOperators {
	m.aggregator.Match(bson.M{field: bson.M{tmorm.ElemMatchOp: condition}})
	return m
}

// And 逻辑与
func (m *MatchOperators) And(conditions ...bson.M) *MatchOperators {
	m.aggregator.Match(bson.M{tmorm.AndOp: conditions})
	return m
}

// Or 逻辑或
func (m *MatchOperators) Or(conditions ...bson.M) *MatchOperators {
	m.aggregator.Match(bson.M{tmorm.OrOp: conditions})
	return m
}

// Not 逻辑非
func (m *MatchOperators) Not(condition bson.M) *MatchOperators {
	m.aggregator.Match(bson.M{tmorm.NotOp: condition})
	return m
}

// Nor 逻辑或非
func (m *MatchOperators) Nor(conditions ...bson.M) *MatchOperators {
	m.aggregator.Match(bson.M{tmorm.NorOp: conditions})
	return m
}

// 时间相关匹配

// DateRange 日期范围
func (m *MatchOperators) DateRange(field string, start, end time.Time) *MatchOperators {
	m.aggregator.Match(bson.M{
		field: bson.M{
			tmorm.GteOp: start,
			tmorm.LteOp: end,
		},
	})
	return m
}

// Today 今天
func (m *MatchOperators) Today(field string) *MatchOperators {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return m.DateRange(field, start, end)
}

// Yesterday 昨天
func (m *MatchOperators) Yesterday(field string) *MatchOperators {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	end := start.Add(24 * time.Hour)
	return m.DateRange(field, start, end)
}

// ThisWeek 本周
func (m *MatchOperators) ThisWeek(field string) *MatchOperators {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日调整为7
	}
	start := now.AddDate(0, 0, -(weekday - 1))
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := start.AddDate(0, 0, 7)
	return m.DateRange(field, start, end)
}

// ThisMonth 本月
func (m *MatchOperators) ThisMonth(field string) *MatchOperators {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return m.DateRange(field, start, end)
}

// LastNDays 最近N天
func (m *MatchOperators) LastNDays(field string, days int) *MatchOperators {
	now := time.Now()
	start := now.AddDate(0, 0, -days)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	return m.DateRange(field, start, now)
}

// Build 构建聚合器
func (m *MatchOperators) Build() *Aggregator[any] {
	return m.aggregator
}

// ProjectOperators 投影操作符
type ProjectOperators struct {
	aggregator *Aggregator[any]
	fields     bson.M
}

// NewProjectOperators 创建投影操作符
func (a *Aggregator[T]) NewProjectOperators() *ProjectOperators {
	return &ProjectOperators{
		aggregator: (*Aggregator[any])(a),
		fields:     make(bson.M),
	}
}

// Include 包含字段
func (p *ProjectOperators) Include(fields ...string) *ProjectOperators {
	for _, field := range fields {
		p.fields[field] = 1
	}
	return p
}

// Exclude 排除字段
func (p *ProjectOperators) Exclude(fields ...string) *ProjectOperators {
	for _, field := range fields {
		p.fields[field] = 0
	}
	return p
}

// ExcludeID 排除_id字段
func (p *ProjectOperators) ExcludeID() *ProjectOperators {
	p.fields["_id"] = 0
	return p
}

// Rename 重命名字段
func (p *ProjectOperators) Rename(oldField, newField string) *ProjectOperators {
	p.fields[newField] = "$" + oldField
	return p
}

// AddField 添加计算字段
func (p *ProjectOperators) AddField(field string, expression any) *ProjectOperators {
	p.fields[field] = expression
	return p
}

// AddConstant 添加常量字段
func (p *ProjectOperators) AddConstant(field string, value any) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.LiteralOp: value}
	return p
}

// Concat 字符串连接
func (p *ProjectOperators) Concat(field string, fields ...string) *ProjectOperators {
	args := make([]any, len(fields))
	for i, f := range fields {
		args[i] = "$" + f
	}
	p.fields[field] = bson.M{tmorm.ConcatOp: args}
	return p
}

// Substr 子字符串
func (p *ProjectOperators) Substr(field, sourceField string, start, length int) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.SubstrBytesOp: []any{"$" + sourceField, start, length}}
	return p
}

// ToLower 转小写
func (p *ProjectOperators) ToLower(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.ToLowerOp: "$" + sourceField}
	return p
}

// ToUpper 转大写
func (p *ProjectOperators) ToUpper(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.ToUpperOp: "$" + sourceField}
	return p
}

// DateToString 日期转字符串
func (p *ProjectOperators) DateToString(field, sourceField, format string, timezone ...string) *ProjectOperators {
	dateToString := bson.M{
		tmorm.FormatOp: format,
		tmorm.DateOp:   "$" + sourceField,
	}
	if len(timezone) > 0 {
		dateToString[tmorm.TimezoneOp] = timezone[0]
	}
	p.fields[field] = bson.M{tmorm.DateToStringOp: dateToString}
	return p
}

// Year 提取年份
func (p *ProjectOperators) Year(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.YearOp: "$" + sourceField}
	return p
}

// Month 提取月份
func (p *ProjectOperators) Month(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.MonthOp: "$" + sourceField}
	return p
}

// Day 提取日期
func (p *ProjectOperators) Day(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.DayOfMonthOp: "$" + sourceField}
	return p
}

// Hour 提取小时
func (p *ProjectOperators) Hour(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.HourOp: "$" + sourceField}
	return p
}

// Add 数学加法
func (p *ProjectOperators) Add(field string, fields ...string) *ProjectOperators {
	args := make([]any, len(fields))
	for i, f := range fields {
		args[i] = "$" + f
	}
	p.fields[field] = bson.M{tmorm.AddOp: args}
	return p
}

// Subtract 数学减法
func (p *ProjectOperators) Subtract(field, field1, field2 string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.SubtractOp: []any{"$" + field1, "$" + field2}}
	return p
}

// Multiply 数学乘法
func (p *ProjectOperators) Multiply(field string, fields ...string) *ProjectOperators {
	args := make([]any, len(fields))
	for i, f := range fields {
		args[i] = "$" + f
	}
	p.fields[field] = bson.M{tmorm.MultiplyOp: args}
	return p
}

// Divide 数学除法
func (p *ProjectOperators) Divide(field, field1, field2 string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.DivideOp: []any{"$" + field1, "$" + field2}}
	return p
}

// Cond 条件表达式
func (p *ProjectOperators) Cond(field string, condition, ifTrue, ifFalse any) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.CondOp: []any{condition, ifTrue, ifFalse}}
	return p
}

// IfNull 空值处理
func (p *ProjectOperators) IfNull(field, sourceField string, replacement any) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.IfNullOp: []any{"$" + sourceField, replacement}}
	return p
}

// Size 数组大小
func (p *ProjectOperators) Size(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.SizeOp: "$" + sourceField}
	return p
}

// Type 字段类型
func (p *ProjectOperators) Type(field, sourceField string) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.TypeOp: "$" + sourceField}
	return p
}

// ArrayElemAt 数组元素
func (p *ProjectOperators) ArrayElemAt(field, sourceField string, index int) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.ArrayElemAtOp: []any{"$" + sourceField, index}}
	return p
}

// First 数组第一个元素
func (p *ProjectOperators) First(field, sourceField string) *ProjectOperators {
	return p.ArrayElemAt(field, sourceField, 0)
}

// Last 数组最后一个元素
func (p *ProjectOperators) Last(field, sourceField string) *ProjectOperators {
	return p.ArrayElemAt(field, sourceField, -1)
}

// Slice 数组切片
func (p *ProjectOperators) Slice(field, sourceField string, start, limit int) *ProjectOperators {
	p.fields[field] = bson.M{tmorm.SliceOp: []any{"$" + sourceField, start, limit}}
	return p
}

// Build 构建投影
func (p *ProjectOperators) Build() *Aggregator[any] {
	return p.aggregator.Project(p.fields)
}

// SortOperators 排序操作符
type SortOperators struct {
	aggregator *Aggregator[any]
	sorts      bson.D
}

// NewSortOperators 创建排序操作符
func (a *Aggregator[T]) NewSortOperators() *SortOperators {
	return &SortOperators{
		aggregator: (*Aggregator[any])(a),
		sorts:      make(bson.D, 0),
	}
}

// Asc 升序
func (s *SortOperators) Asc(fields ...string) *SortOperators {
	for _, field := range fields {
		s.sorts = append(s.sorts, bson.E{Key: field, Value: 1})
	}
	return s
}

// Desc 降序
func (s *SortOperators) Desc(fields ...string) *SortOperators {
	for _, field := range fields {
		s.sorts = append(s.sorts, bson.E{Key: field, Value: -1})
	}
	return s
}

// By 按字段排序
func (s *SortOperators) By(field string, order int) *SortOperators {
	s.sorts = append(s.sorts, bson.E{Key: field, Value: order})
	return s
}

// TextScore 文本搜索分数排序
func (s *SortOperators) TextScore() *SortOperators {
	s.sorts = append(s.sorts, bson.E{Key: "score", Value: bson.M{tmorm.MetaOp: "textScore"}})
	return s
}

// Build 构建排序
func (s *SortOperators) Build() *Aggregator[any] {
	return s.aggregator.Sort(s.sorts)
}

// LookupOperators 关联查询操作符
type LookupOperators struct {
	aggregator *Aggregator[any]
}

// NewLookupOperators 创建关联查询操作符
func (a *Aggregator[T]) NewLookupOperators() *LookupOperators {
	return &LookupOperators{
		aggregator: (*Aggregator[any])(a),
	}
}

// Simple 简单关联
func (l *LookupOperators) Simple(from, localField, foreignField, as string) *LookupOperators {
	l.aggregator.Lookup(from, localField, foreignField, as)
	return l
}

// WithPipeline 带管道的关联
func (l *LookupOperators) WithPipeline(from, as string, let bson.M, pipeline []bson.M) *LookupOperators {
	lookupStage := bson.M{
		"from": from,
		"as":   as,
	}
	if let != nil {
		lookupStage["let"] = let
	}
	if pipeline != nil {
		lookupStage["pipeline"] = pipeline
	}
	l.aggregator.AddStage(bson.M{tmorm.StageLookUpOp: lookupStage})
	return l
}

// Build 构建关联查询
func (l *LookupOperators) Build() *Aggregator[any] {
	return l.aggregator
}

// UnwindOperators 展开操作符
type UnwindOperators struct {
	aggregator *Aggregator[any]
}

// NewUnwindOperators 创建展开操作符
func (a *Aggregator[T]) NewUnwindOperators() *UnwindOperators {
	return &UnwindOperators{
		aggregator: (*Aggregator[any])(a),
	}
}

// Simple 简单展开
func (u *UnwindOperators) Simple(path string) *UnwindOperators {
	u.aggregator.Unwind(path)
	return u
}

// WithOptions 带选项的展开
func (u *UnwindOperators) WithOptions(path string, preserveNullAndEmptyArrays bool, includeArrayIndex string) *UnwindOperators {
	unwindStage := bson.M{"path": "$" + path}
	if preserveNullAndEmptyArrays {
		unwindStage["preserveNullAndEmptyArrays"] = true
	}
	if includeArrayIndex != "" {
		unwindStage["includeArrayIndex"] = includeArrayIndex
	}
	u.aggregator.AddStage(bson.M{tmorm.UnwindOp: unwindStage})
	return u
}

// Build 构建展开
func (u *UnwindOperators) Build() *Aggregator[any] {
	return u.aggregator
}

// 便利方法

// MatchWhere 匹配条件
func (a *Aggregator[T]) MatchWhere() *MatchOperators {
	return a.NewMatchOperators()
}

// ProjectSelect 投影选择
func (a *Aggregator[T]) ProjectSelect() *ProjectOperators {
	return a.NewProjectOperators()
}

// SortOrder 排序
func (a *Aggregator[T]) SortOrder() *SortOperators {
	return a.NewSortOperators()
}

// LookupJoin 关联查询
func (a *Aggregator[T]) LookupJoin() *LookupOperators {
	return a.NewLookupOperators()
}

// UnwindArray 展开数组
func (a *Aggregator[T]) UnwindArray() *UnwindOperators {
	return a.NewUnwindOperators()
}

// 高级聚合操作

// GraphLookup 图查找
func (a *Aggregator[T]) GraphLookup(from, startWith, connectFromField, connectToField, as string, maxDepth ...int) *Aggregator[T] {
	graphLookup := bson.M{
		"from":             from,
		"startWith":        startWith,
		"connectFromField": connectFromField,
		"connectToField":   connectToField,
		"as":               as,
	}
	if len(maxDepth) > 0 {
		graphLookup["maxDepth"] = maxDepth[0]
	}
	return a.AddStage(bson.M{tmorm.GraphLookupOp: graphLookup})
}

// Sample 随机采样
func (a *Aggregator[T]) Sample(size int) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.SampleOp: bson.M{"size": size}})
}

// Out 输出到集合
func (a *Aggregator[T]) Out(collection string) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.OutOp: collection})
}

// Merge 合并到集合
func (a *Aggregator[T]) Merge(into any, whenMatched, whenNotMatched string) *Aggregator[T] {
	mergeStage := bson.M{"into": into}
	if whenMatched != "" {
		mergeStage["whenMatched"] = whenMatched
	}
	if whenNotMatched != "" {
		mergeStage["whenNotMatched"] = whenNotMatched
	}
	return a.AddStage(bson.M{tmorm.MergeOp: mergeStage})
}

// Redact 文档过滤
func (a *Aggregator[T]) Redact(expression any) *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.RedactOp: expression})
}

// GeoNear 地理位置查询
func (a *Aggregator[T]) GeoNear(near any, distanceField string, options bson.M) *Aggregator[T] {
	geoNear := bson.M{
		"near":          near,
		"distanceField": distanceField,
	}
	for k, v := range options {
		geoNear[k] = v
	}
	return a.AddStage(bson.M{tmorm.GeoNearOp: geoNear})
}

// IndexStats 索引统计
func (a *Aggregator[T]) IndexStats() *Aggregator[T] {
	return a.AddStage(bson.M{tmorm.IndexStatsOp: bson.M{}})
}

// CollStats 集合统计
func (a *Aggregator[T]) CollStats(options bson.M) *Aggregator[T] {
	if options == nil {
		options = bson.M{}
	}
	return a.AddStage(bson.M{tmorm.CollStatsOp: options})
}

// ListSessions 列出会话
func (a *Aggregator[T]) ListSessions(options bson.M) *Aggregator[T] {
	if options == nil {
		options = bson.M{}
	}
	return a.AddStage(bson.M{tmorm.ListSessionsOp: options})
}

// CurrentOp 当前操作
func (a *Aggregator[T]) CurrentOp(options bson.M) *Aggregator[T] {
	if options == nil {
		options = bson.M{}
	}
	return a.AddStage(bson.M{tmorm.CurrentOpOp: options})
}

// 文本搜索

// TextSearch 文本搜索
func (a *Aggregator[T]) TextSearch(search string, language ...string) *Aggregator[T] {
	textSearch := bson.M{tmorm.SearchOp: search}
	if len(language) > 0 {
		textSearch[tmorm.LanguageOp] = language[0]
	}
	return a.Match(bson.M{tmorm.TextOp: textSearch})
}

// 数据类型转换

// Convert 类型转换
func (a *Aggregator[T]) Convert(input any, to string, onError, onNull any) *AggExpr {
	convert := bson.M{
		tmorm.InputOp: input,
		"to":          to,
	}
	if onError != nil {
		convert["onError"] = onError
	}
	if onNull != nil {
		convert[tmorm.OnNullOp] = onNull
	}
	return &AggExpr{expr: bson.M{tmorm.ConvertOp: convert}}
}

// ToBool 转换为布尔值
func (a *Aggregator[T]) ToBool(input any) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToBoolOp: input}}
}

// ToDecimal 转换为十进制
func (a *Aggregator[T]) ToDecimal(input any) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToDecimalOp: input}}
}

// ToLong 转换为长整型
func (a *Aggregator[T]) ToLong(input any) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToLongOp: input}}
}

// ToObjectId 转换为ObjectId
func (a *Aggregator[T]) ToObjectId(input any) *AggExpr {
	return &AggExpr{expr: bson.M{tmorm.ToObjectIdOp: input}}
}
