<p align="right">
   <strong>中文</strong> | <a href="./README.md">English</a>
</p>

# tmORM

tmORM 是一个轻量级的 Go MongoDB ORM。它在官方 `mongo-driver` 之上提供了类型安全的链式查询构建器，以及更新构建器、
聚合管道构建器，和一套用于处理日志、重试、熔断等横切关注点的中间件机制。

本文档中的所有示例都经过实际编译验证。

## 特性

- 类型安全的泛型查询构建器（`Collection[T]`）
- 可组合的查询条件：链式 `Where`、表达式树、原生 BSON
- 更新构建器，覆盖 `$set` / `$inc` / `$mul` / 数组操作符 / `$rename` / `$currentDate`
- 聚合管道构建器，结果带类型
- Repository 模式，内置常用 CRUD 封装
- 中间件链，内置日志、panic 恢复、耗时监控、重试、错误处理、熔断
- 结构化结果与错误分类（`MResult`）
- 多客户端管理与事务封装

## 环境要求

- Go 1.21+
- MongoDB 4.2+（事务需要副本集）

## 安装

```bash
go get github.com/zzjha-cn/tm_orm
```

## 快速开始

### 连接

```go
import (
    "time"

    tmorm "github.com/zzjha-cn/tm_orm"
)

config := tmorm.DefaultClientConfig("mongodb://localhost:27017")
config.MaxPoolSize = 100
config.ConnectTimeout = 10 * time.Second

ormClient, err := tmorm.NewORMClient(config)
if err != nil {
    panic(err)
}
defer ormClient.Close()
```

`NewORMClient` 会立即建立连接并 ping，因此返回值非 nil 时即可直接使用。config 传 `nil` 时默认连接
`mongodb://localhost:27017`。

### 定义模型并查询

```go
import (
    "github.com/zzjha-cn/tm_orm/collection"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Age   int                `bson:"age"`
    Email string             `bson:"email"`
}

coll := collection.NewCollection[User](ormClient, "mydb", "users")

// 查询多条
users, err := coll.Where("name").Eq("sean").Where("age").Gte(18).Find(ctx)

// 查询单条
user, err := coll.Where("_id").Eq(userID).FindOne(ctx)

// 计数
count, err := coll.Where("age").Gte(18).Count(ctx)

// 原生 driver options 直接透传
opts := options.Find().SetSort(bson.D{{Key: "age", Value: -1}}).SetLimit(10)
users, err = coll.Where("age").Gte(18).Find(ctx, opts)
```

## 查询

### 链式条件

链上的每个 `Where(...)` 都以 `$and` 组合，因此同一个字段可以重复出现：

```go
users, err := coll.
    Where("age").Gte(18).Where("age").Lte(65).
    Where("name").Regex("^[A-Z]").
    Where("email").Exists(true).
    Find(ctx)
```

生成：

```js
{ $and: [ { age: { $gte: 18 } }, { age: { $lte: 65 } },
          { name: { $regex: "^[A-Z]" } }, { email: { $exists: true } } ] }
```

`Collection` 是不可变的：每加一个条件都会返回新的实例，因此可以安全地把一个已带条件的集合作为多个查询的基础复用。

`Where(field)` 支持：`Eq`、`Ne`、`Gt`、`Gte`、`Lt`、`Lte`、`In`、`NotIn`、`Regex`、`Exists`，以及用 `Not()` 对下一个
操作符取反。

```go
// 集合成员
users, err := coll.Where("city").In("Beijing", "Shanghai", "Guangzhou").Find(ctx)

// 取反 -> { status: { $not: { $eq: "active" } } }
users, err = coll.Where("status").Not().Eq("active").Find(ctx)
```

### 原生 BSON 条件

`Filter` 和 `AndFilter` 接受任意实现了 `impl.IBsonQuery` 的类型，该接口只有一个方法：

```go
import "github.com/zzjha-cn/tm_orm/impl"

type RawQuery struct {
    Data bson.D
}

func (q *RawQuery) GetBsonD() bson.D { return q.Data }

query := &RawQuery{Data: bson.D{
    {Key: "$or", Value: bson.A{
        bson.D{{Key: "name", Value: "sean"}},
        bson.D{{Key: "email", Value: bson.M{"$regex": "@example.com$"}}},
    }},
}}

// Filter 会覆盖此前已有的条件
users, err := coll.Filter(query).Find(ctx)

// AndFilter 则以 $and 追加
users, err = coll.Where("age").Gte(18).AndFilter(query).Find(ctx)
```

### 表达式

`Expr(field)` 比 `Where` 提供了更丰富的操作符，包含字符串、范围、数组和日期相关的便捷方法：

```go
users, err := coll.Expr("age").Between(18, 65).Find(ctx)
users, err = coll.Expr("email").Contains("@company.com").Find(ctx)
users, err = coll.Expr("created_at").LastNDays(7).Find(ctx)
users, err = coll.Expr("tags").All("golang", "backend").Find(ctx)

// 每个 Expr(...) 的终结方法都返回 Collection，因此可以继续以 $and 串联
users, err = coll.Expr("age").Gt(25).Expr("status").Eq("active").Find(ctx)
```

`Expr(field)` 支持 `Eq`、`Ne`、`Gt`、`Gte`、`Lt`、`Lte`、`In`、`NotIn`、`Exists`、`Regex`、`Contains`、
`StartsWith`、`EndsWith`、`Between`、`IsNull`、`IsNotNull`、`Size`、`All`、`ElemMatch`，以及日期快捷方法
`Today`、`Yesterday`、`ThisWeek`、`ThisMonth`、`LastNDays`。

需要布尔逻辑时，用 `expression` 包构建表达式，再用 `And` / `Or` / `Nor` / `Not` 组合：

```go
import "github.com/zzjha-cn/tm_orm/expression"

expr1 := expression.NewExpression().Field("name").Eq("sean")
expr2 := expression.NewExpression().Field("age").Gt(25)

users, err := coll.WhereExpr(expression.Or(expr1, expr2)).Find(ctx)
users, err = coll.WhereExpr(expression.And(expr1, expr2)).Find(ctx)
users, err = coll.WhereExpr(expression.Not(expr1)).Find(ctx)

// expression.Q 是 NewExpression().Field(...) 的简写
nested := expression.And(
    expression.Or(
        expression.Q("department").Eq("Engineering"),
        expression.Q("department").Eq("IT"),
    ),
    expression.Q("age").Between(25, 45),
)
users, err = coll.WhereExpr(nested).Find(ctx)
```

> **注意：** 单个 `Expression` 只持有一个逻辑操作符。在它上面调用 `.And()` / `.Or()` 会把**全部**条件的逻辑操作符
> 一起改掉，并不能混用。需要同时使用 AND 和 OR 时，请按上面的方式嵌套 `expression.And(...)` 与
> `expression.Or(...)`。`WhereExpr` 本身以 `$and` 追加，因此也可以和 `Where` 混用。

## 更新

```go
// 单字段
res, err := coll.Where("_id").Eq(userID).Set("age", 25).UpdateOne(ctx)

// 多字段，更新所有匹配文档
res, err = coll.Where("name").Eq("sean").
    Set("age", 26).
    Set("email", "sean@example.com").
    Update(ctx)

// 数值操作
res, err = coll.Where("_id").Eq(userID).
    Inc("age", 1).
    Mul("salary", 1.1).
    UpdateOne(ctx)

// 删除字段 / 重命名 / 时间戳
res, err = coll.Where("_id").Eq(userID).
    Unset("deprecated_field").
    Rename("old_field", "new_field").
    CurrentDate("updated_at").
    UpdateOne(ctx)

// upsert
res, err = coll.Where("name").Eq("sean").
    Set("age", 30).
    SetOnInsert("created_at", time.Now()).
    UpdateOne(ctx, options.Update().SetUpsert(true))
```

`Collection` 上可以直接起链的只有 `Set`、`Inc`、`Unset` 三个方法。其余操作符（包括所有数组操作）都定义在
`UpdateBuilder` 上，因此要么先用这三个之一起链，要么显式构造 builder：

```go
res, err := collection.NewUpdateBuilder(coll.Where("_id").Eq(userID)).
    Push("tags", "new-tag").
    Pull("old_tags", "deprecated-tag").
    AddToSet("skills", "golang").
    UpdateOne(ctx)
```

`UpdateOne` 对应 `UpdateOne`，`Update` 对应 `UpdateMany`。

> **注意：** 不带任何条件的 `Update` 和 `Delete` 会命中整个集合。除非确实需要，否则务必先设置过滤条件。

## 插入与删除

```go
user := &User{Name: "sean", Age: 25, Email: "sean@example.com"}
one, err := coll.Insert(ctx, user)

users := []*User{
    {Name: "alice", Age: 23, Email: "alice@example.com"},
    {Name: "bob", Age: 27, Email: "bob@example.com"},
}
many, err := coll.InsertMany(ctx, users)

del, err := coll.Where("_id").Eq(user.ID).DeleteOne(ctx)
del, err = coll.Where("age").Lt(18).Delete(ctx) // DeleteMany
```

## Repository

```go
import "github.com/zzjha-cn/tm_orm/repo"

r := repo.NewRepository[User](ormClient, "mydb", "users")

created, err := r.Create(ctx, &User{Name: "sean", Age: 25})   // *mongo.InsertOneResult
id := created.InsertedID

found, err := r.FindByID(ctx, id)                             // *User
updateRes, err := r.UpdateByID(ctx, id, map[string]any{       // *mongo.UpdateResult
    "age":   26,
    "email": "newemail@example.com",
})
deleteRes, err := r.DeleteByID(ctx, id)                       // *mongo.DeleteResult

list, total, err := r.FindPage(ctx, 1, 10)                    // 第 1 页，每页 10 条
exists, err := r.Exists(ctx, id)
count, err := r.Count(ctx)
```

`BaseRepository` 在此之上补充了按字段操作的封装：

```go
br := repo.NewBaseRepository[User](ormClient, "mydb", "users")

byField, err := br.FindByField(ctx, "name", "sean")
one, err := br.FindOneByField(ctx, "email", "sean@example.com")
in, err := br.FindIn(ctx, "city", "Beijing", "Shanghai")
n, err := br.CountByField(ctx, "status", "active")
```

注意 `FindPage` 统计的是整个集合的总数，而不是过滤后的子集。

## 聚合

`Aggregator[T]` 会把结果解码为 `T`，所以泛型参数应当是**输出结果的结构**，而不是源文档：

```go
import "github.com/zzjha-cn/tm_orm/aggregator"

type DeptStat struct {
    Department string  `bson:"_id"`
    AvgAge     float64 `bson:"avgAge"`
    Count      int64   `bson:"count"`
}

agg := aggregator.NewAggregator[DeptStat](ormClient, "mydb", "users")

stats, err := agg.
    Match(bson.M{"age": bson.M{"$gte": 18}}).
    Group("$department", bson.M{
        "avgAge": bson.M{"$avg": "$age"},
        "count":  bson.M{"$sum": 1},
    }).
    SortDesc("avgAge").
    Execute(ctx)
```

`$match` 也可以用表达式构建：

```go
stats, err := agg.
    MatchByExpression(expression.Q("status").Eq("active")).
    Group("$department", bson.M{"count": bson.M{"$sum": 1}}).
    Execute(ctx)
```

### 关联查询

```go
type OrderResult struct {
    ID        primitive.ObjectID `bson:"_id"`
    Amount    float64            `bson:"amount"`
    UserName  string             `bson:"user_name"`
    UserEmail string             `bson:"user_email"`
}

orderAgg := aggregator.NewAggregator[OrderResult](ormClient, "mydb", "orders")

orders, err := orderAgg.
    Match(bson.M{"status": "completed"}).
    Lookup("users", "user_id", "_id", "user_info").
    Unwind("user_info").                 // 前面的 "$" 会自动补上
    Project(bson.M{
        "_id":        1,
        "amount":     1,
        "user_name":  "$user_info.name",
        "user_email": "$user_info.email",
    }).
    Execute(ctx)

// 保留数组为空或字段缺失的文档
orders, err = orderAgg.Unwind("user_info", true).Execute(ctx)
```

### 其他阶段

```go
// 分面聚合
facet, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "users").
    Match(bson.M{"status": "active"}).
    Facet(bson.M{
        "ageStats": []bson.M{
            {"$group": bson.M{"_id": nil, "avgAge": bson.M{"$avg": "$age"}}},
        },
        "departmentCount": []bson.M{
            {"$group": bson.M{"_id": "$department", "count": bson.M{"$sum": 1}}},
            {"$sort": bson.M{"count": -1}},
        },
    }).
    Execute(ctx)

// 分页：Paginate 只负责拼 $skip/$limit，仍需自行调用 Execute
page, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "users").
    Match(bson.M{"age": bson.M{"$gte": 18}}).
    SortAsc("name").
    Paginate(1, 10).
    Execute(ctx)

// $replaceRoot 接受 { newRoot: <表达式> }；$replaceWith 直接接受表达式
rooted, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "orders").
    Unwind("items").
    ReplaceRoot("$items").
    Execute(ctx)

withed, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "orders").
    ReplaceWith("$items").
    Execute(ctx)
```

`ExecuteOne` 返回第一条结果（无结果时返回 `mongo.ErrNoDocuments`），`ExecuteRaw` 返回原生 cursor：

```go
single, err := agg.Match(bson.M{"name": "sean"}).ExecuteOne(ctx)

cursor, err := agg.Match(bson.M{"name": "sean"}).ExecuteRaw(ctx)
if err == nil {
    defer cursor.Close(ctx)
}
```

> **注意：** 与 `Collection` 不同，`Aggregator` 是可变的——每个阶段都会追加到同一条管道上。请为每次查询新建一个
> aggregator，另外注意 `ExecuteOne` 会往管道里追加一个 `$limit` 阶段并且不会移除。

## 中间件

中间件是一个接收操作上下文和后续链路的函数。执行顺序为：客户端级中间件 → 集合级中间件 → 实际的数据库调用。

```go
loggingMiddleware := func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
    start := time.Now()
    fmt.Printf("start: %s %s.%s\n", mctx.Typ, mctx.DBName, mctx.CollectionName)

    next(mctx)

    if mctx.Result.Err != nil {
        fmt.Printf("failed after %v: %v\n", time.Since(start), mctx.Result.Err)
        return
    }
    fmt.Printf("done in %v\n", time.Since(start))
}

// 客户端级别：对该客户端的所有操作生效
ormClient.Use(
    tmorm.Recovery(),
    tmorm.LoggingMiddleware(log.Default()),
    tmorm.TimingMiddleware(100*time.Millisecond),
    tmorm.RetryMiddleware(3, 200*time.Millisecond),
    tmorm.CircuitBreakerMiddleware(5, 30*time.Second),
)

// 集合级别：只对该集合生效
coll := collection.NewCollection[User](ormClient, "mydb", "users", loggingMiddleware)
```

内置中间件：

| 中间件 | 作用 |
| --- | --- |
| `Recovery()` | 捕获 panic，转为带堆栈的 fatal `MResult` |
| `LoggingMiddleware(logger)` | 记录操作类型、过滤条件、耗时与错误详情 |
| `TimingMiddleware(threshold)` | 把慢操作标记为 warning |
| `RetryMiddleware(max, delay)` | 对可重试错误（网络 / 超时）进行重试 |
| `ErrorHandlingMiddleware(cb)` | 用分类后的错误信息回调 |
| `CircuitBreakerMiddleware(n, reset)` | 连续失败 `n` 次后熔断 |

在中间件内部，`mctx.Operation` 是过滤条件，`mctx.UpdateOp` 是更新文档，`mctx.AggOp` 是聚合管道，
`mctx.Result` 是执行结果。

### 错误处理

每次操作都会填充一个 `MResult`，错误会被自动分类为类型、错误码和是否可重试：

```go
callback := func(result *tmorm.MResult, mctx *tmorm.MiddleCtx) {
    fmt.Printf("operation=%s type=%s code=%s retryable=%v err=%v\n",
        mctx.Typ, result.ErrorType, result.Code, result.Retryable, result.Err)

    if result.Level == tmorm.ErrorLevelFatal {
        fmt.Println("fatal:", result.GetErrorInfo())
    }
}

ormClient.Use(tmorm.ErrorHandlingMiddleware(callback))
```

错误类型：`CONNECTION`、`TIMEOUT`、`VALIDATION`、`QUERY`、`UPDATE`、`INSERT`、`DELETE`、`AGGREGATE`、`AUTH`、
`INDEX`、`UNKNOWN`。错误级别：`INFO`、`WARNING`、`ERROR`、`FATAL`。

## 事务

事务需要副本集或分片集群。事务内的每个操作都要传入 session context：

```go
err := ormClient.Transaction(ctx, func(sc mongo.SessionContext) error {
    coll := collection.NewCollection[User](ormClient, "mydb", "users")

    if _, err := coll.Insert(sc, &User{Name: "sean", Age: 25}); err != nil {
        return err
    }

    _, err := coll.Where("name").Eq("alice").Inc("age", 1).UpdateOne(sc)
    return err
})
```

## 多客户端

```go
manager := tmorm.NewClientManager()
defer manager.Close()

_ = manager.AddClient("primary", tmorm.DefaultClientConfig("mongodb://localhost:27017"))
_ = manager.AddClient("analytics", tmorm.DefaultClientConfig("mongodb://localhost:27018"))
_ = manager.SetDefaultClient("primary")

primary, err := manager.GetClient("primary")

for name, err := range manager.HealthCheck(ctx) {
    fmt.Println(name, err)
}
```

也可以通过 `tmorm.GlobalManager()`、`tmorm.AddGlobalClient(...)`、`tmorm.GetGlobalClient(...)` 使用进程级别的
全局管理器。

## 支持的操作符

- 比较：`$eq`、`$gt`、`$gte`、`$lt`、`$lte`、`$ne`、`$in`、`$nin`
- 逻辑：`$and`、`$or`、`$not`、`$nor`
- 元素：`$exists`、`$type`
- 求值：`$expr`、`$regex`、`$mod`、`$text`
- 数组：`$all`、`$elemMatch`、`$size`
- 更新：`$set`、`$unset`、`$inc`、`$mul`、`$rename`、`$setOnInsert`、`$currentDate`、`$push`、`$pop`、`$pull`、
  `$pullAll`、`$addToSet`
- 聚合阶段：`$match`、`$group`、`$sort`、`$project`、`$limit`、`$skip`、`$unwind`、`$lookup`、`$graphLookup`、
  `$addFields`、`$replaceRoot`、`$replaceWith`、`$facet`、`$bucket`、`$sample`、`$out`、`$merge`、`$redact`、
  `$geoNear`、`$count`

完整列表详见 [docs/support_exprssion.md](docs/support_exprssion.md)。

## 已知限制

以下几点建议在使用前了解：

- 单个 `Expression` 只持有一个逻辑操作符；需要混用 AND 和 OR 时请嵌套 `expression.And` / `expression.Or`，
  而不是在同一个表达式上链式调用 `.And()` / `.Or()`。
- `Aggregator` 会就地修改自己的管道，请为每次查询新建一个。
- `Aggregator.PaginateWithCount` 尚未实现，请用 `Paginate` 配合单独的计数查询。
- `GroupBuilder` 以及 `MatchOperators` / `ProjectOperators` / `SortOperators` 这几个辅助构建器的 `Build()`
  返回的是 `Aggregator[any]`，会丢失结果类型。需要类型化结果时请直接用 `Group(id, fields)` 和 `Match(filter)`。
- 不带过滤条件的 `Update` 和 `Delete` 会作用于整个集合。
- `Repository.FindPage` 返回的是整个集合的总数，而不是过滤后的总数。
- 中间件目前无法替换驱动调用实际使用的 `context.Context`。

## 进度

- [x] 完成 query builder
- [x] TDD 测试驱动
- [x] 完成 query expr、and、or
- [x] 完成 query aggregate operator 聚合操作符
- [x] 完成 Middleware
- [x] 简单完成 Aggregator
- [x] 完成端到端测试
- [x] 查询的错误处理增强，支持错误分类与错误回调
- [ ] 完成 Agg stage 阶段的各个常用命令封装
- [ ] 增加元数据管理中心
- [ ] 增加原子化操作
- [ ] 增加结果集处理封装

## 性能测试

```
goos: darwin
goarch: arm64
pkg: github.com/zzjha-cn/tm_orm/test
BenchmarkBasicQuery
BenchmarkBasicQuery/ORM-SimpleQuery
BenchmarkBasicQuery/ORM-SimpleQuery-8         	      12	  91372882 ns/op	14673922 B/op	  390403 allocs/op
BenchmarkBasicQuery/Native-SimpleQuery
BenchmarkBasicQuery/Native-SimpleQuery-8      	      13	  89366000 ns/op	21671820 B/op	  380370 allocs/op
BenchmarkComplexQuery
BenchmarkComplexQuery/ORM-ComplexQuery
BenchmarkComplexQuery/ORM-ComplexQuery-8      	      12	  94856938 ns/op	14675226 B/op	  390509 allocs/op
BenchmarkComplexQuery/Native-ComplexQuery
BenchmarkComplexQuery/Native-ComplexQuery-8   	      14	  76689161 ns/op	16657950 B/op	  296384 allocs/op
BenchmarkUpdate
BenchmarkUpdate/ORM-Update
BenchmarkUpdate/ORM-Update-8                  	     295	  11141763 ns/op	    7690 B/op	     128 allocs/op
BenchmarkUpdate/Native-Update
BenchmarkUpdate/Native-Update-8               	     306	   4011816 ns/op	    6079 B/op	      91 allocs/op
BenchmarkInsert
BenchmarkInsert/ORM-InsertOne
BenchmarkInsert/ORM-InsertOne-8               	   13258	     88546 ns/op	    7363 B/op	     123 allocs/op
BenchmarkInsert/Native-InsertOne
BenchmarkInsert/Native-InsertOne-8            	   13843	     84566 ns/op	    6046 B/op	      88 allocs/op
BenchmarkInsert/ORM-InsertMany
BenchmarkInsert/ORM-InsertMany-8              	     889	   3823420 ns/op	  255304 B/op	    4480 allocs/op
BenchmarkJoinQuery
BenchmarkJoinQuery/ORM-LookupJoin
BenchmarkJoinQuery/ORM-LookupJoin-8           	      19	  62841002 ns/op	 2022170 B/op	   39921 allocs/op
BenchmarkJoinQuery/Native-LookupJoin
BenchmarkJoinQuery/Native-LookupJoin-8        	      16	  63172367 ns/op	 2445467 B/op	   53703 allocs/op
BenchmarkRepository
BenchmarkRepository/Repository-FindByID
BenchmarkRepository/Repository-FindByID-8     	   12176	     97524 ns/op	    9287 B/op	     142 allocs/op
BenchmarkRepository/Repository-FindPage
BenchmarkRepository/Repository-FindPage-8     	     532	   5194989 ns/op	   40738 B/op	     943 allocs/op
BenchmarkRepository/Repository-UpdateByID
BenchmarkRepository/Repository-UpdateByID-8   	   10280	    110071 ns/op	    7657 B/op	     132 allocs/op
BenchmarkConcurrent
BenchmarkConcurrent/ORM-ConcurrentRead
BenchmarkConcurrent/ORM-ConcurrentRead-8      	      26	  62581518 ns/op	39395269 B/op	 1025371 allocs/op
BenchmarkConcurrent/ORM-ConcurrentWrite
BenchmarkConcurrent/ORM-ConcurrentWrite-8     	   35160	     33733 ns/op	    7372 B/op	     123 allocs/op
BenchmarkConcurrent/ORM-ConcurrentUpdate
BenchmarkConcurrent/ORM-ConcurrentUpdate-8    	     177	   7399120 ns/op	    7634 B/op	     129 allocs/op
BenchmarkMemoryUsage
BenchmarkMemoryUsage/ORM-LargeResultSet
BenchmarkMemoryUsage/ORM-LargeResultSet-8     	       2	 687059896 ns/op	136147088 B/op	 3514824 allocs/op
BenchmarkMemoryUsage/ORM-StreamingQuery
BenchmarkMemoryUsage/ORM-StreamingQuery-8     	       8	 137702141 ns/op	14462405 B/op	  392079 allocs/op
BenchmarkComplexConditions
BenchmarkComplexConditions/ORM-MultipleConditions
BenchmarkComplexConditions/ORM-MultipleConditions-8         	       3	 335019486 ns/op	56357880 B/op	 1458175 allocs/op
BenchmarkComplexConditions/Native-MultipleConditions
BenchmarkComplexConditions/Native-MultipleConditions-8      	       6	 205371618 ns/op	45832770 B/op	  779018 allocs/op
BenchmarkComplexConditions/ORM-RegexQuery
BenchmarkComplexConditions/ORM-RegexQuery-8                 	       3	 382153375 ns/op	75060738 B/op	 1956364 allocs/op
BenchmarkIndexPerformance
BenchmarkIndexPerformance/ORM-IndexedQuery
BenchmarkIndexPerformance/ORM-IndexedQuery-8                	      44	  32730887 ns/op	 4804644 B/op	  130477 allocs/op
BenchmarkIndexPerformance/ORM-NonIndexedQuery
BenchmarkIndexPerformance/ORM-NonIndexedQuery-8             	      24	  47871872 ns/op	    8132 B/op	     105 allocs/op
BenchmarkIndexPerformance/ORM-CompoundIndexQuery
BenchmarkIndexPerformance/ORM-CompoundIndexQuery-8          	       1	1599281500 ns/op	240992136 B/op	 6243937 allocs/op
BenchmarkDelete
BenchmarkDelete/ORM-DeleteMany
BenchmarkDelete/ORM-DeleteMany-8                            	   12594	     96674 ns/op	    5843 B/op	      95 allocs/op
BenchmarkDelete/Native-DeleteMany
BenchmarkDelete/Native-DeleteMany-8                         	   13174	     94205 ns/op	    4384 B/op	      60 allocs/op
BenchmarkDelete/ORM-DeleteOne
BenchmarkDelete/ORM-DeleteOne-8                             	    8694	    126434 ns/op	    5842 B/op	      95 allocs/op
BenchmarkPagination
BenchmarkPagination/ORM-Pagination
BenchmarkPagination/ORM-Pagination-8                        	     498	   2897570 ns/op	   38169 B/op	     931 allocs/op
BenchmarkPagination/Native-Pagination
BenchmarkPagination/Native-Pagination-8                     	     438	   3315105 ns/op	   45480 B/op	     872 allocs/op
BenchmarkCount
BenchmarkCount/ORM-Count
BenchmarkCount/ORM-Count-8                                  	      58	  18980491 ns/op	    7609 B/op	     118 allocs/op
BenchmarkCount/Native-Count
BenchmarkCount/Native-Count-8                               	      13	  82383907 ns/op	    6720 B/op	      82 allocs/op
BenchmarkAggregate
BenchmarkAggregate/ORM-SimpleAggregate
BenchmarkAggregate/Native-SimpleAggregate
BenchmarkAggregate/Native-SimpleAggregate-8                 	      24	  48347632 ns/op	   65619 B/op	    1276 allocs/op
BenchmarkAggregate/ORM-ComplexAggregate
BenchmarkAggregate/ORM-ComplexAggregate-8                   	      21	  51887315 ns/op	   25726 B/op	     417 allocs/op
BenchmarkAggregate/Native-ComplexAggregate
BenchmarkAggregate/Native-ComplexAggregate-8                	      21	  52452042 ns/op	   21222 B/op	     439 allocs/op
```

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。
