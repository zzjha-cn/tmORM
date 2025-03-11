# tmORM

tmORM是一个轻量级的MongoDB ORM框架，提供了简洁而强大的API来操作MongoDB数据库。它支持链式调用、类型安全的查询构建器、更新操作和聚合表达式等特性。

## 特性

- 类型安全的查询构建器
- 支持复杂的查询条件（AND/OR/表达式）
- 支持文档更新和替换操作
- 支持数组操作（Push/Pull/AddToSet）
- 支持Upsert操作
- 支持聚合管道操作
- 内置中间件机制，支持自定义中间件
- 完整的查询表达式支持

## 安装

```bash
go get
```

## 快速开始
### 连接数据库
```go
import (
    "context"
    tmorm "tm_orm"
    "go.mongodb.org/mongo-driver/mongo"
)

// 创建 MongoDB 客户端
client, err := mongo.Connect(context.Background())
if err != nil {
    panic(err)
}

// 初始化数据库管理器
db := tmorm.NewMDB(client)
```
### 基本查询

```go
import (
    "tm_orm/finder"
    "tm_orm/query"
)

// 定义模型
type User struct {
    ID   primitive.ObjectID `bson:"_id"`
    Name string            `bson:"name"`
    Age  int64             `bson:"age"`
}

// 创建查询器
fd := &finder.Finder[User]{}

// 构建查询条件
filter := query.Query{}.Builder().
    K("name").Eq("sean").
    K("age").Gte(18).
    ToQuery()

// 执行查询
users, err := fd.Find(
    db.Sess(context.Background(), "mydb", "users"),
    filter,
)
```

### 复杂查询

#### AND/OR条件
```go
filter := query.Query{}.Builder().
    Or(func(a *query.QueryOr) query.Builder {
        return a.
            K("age").Gt(10).
            K("name").Eq("xx").
            And(func(and *query.QueryAnd) query.Builder {
                return and.
                    K("age").Gt(100).
                    K("name").Eq("sean")
            })
    }).ToQuery()
```

#### 表达式查询
```go
filter := query.Query{}.Builder().
    Expr(func(m query.MExpr) query.Builder {
        return m.Or(
            m.C().Eq(m.Fd("name"), "sean2"),
            m.Er().And(
                m.C().Eq(
                    m.C().Multi(m.Fd("age"), 2),
                    40,
                ),
                m.C().In(m.Fd("name"), "sean"),
            ),
        )
    }).ToQuery()
```

### 更新操作

```go
import "tm_orm/updater"

// 创建更新器
up := &updater.MUpdater[User]{}

// 单字段更新
bd := updater.NewReplaceBuilder[User]()
bd.C().Set("age", 25)

// 执行更新
_, err := up.SetFilter(
    query.Query{}.Builder().K("_id").Eq(id).ToQuery(),
).UpdateOne(db.Sess(context.Background(), "mydb", "users"), bd)

// 对象更新
user := &User{ID: id, Name: "sean", Age: 20}
bd := updater.NewReplaceBuilder[User]()
bd.C().SetObj(user, false)

// 数组操作
bd.C().
    AddToSet("tags", []string{"tag1", "tag2"}).
    Push("scores", []int{90, 95}).
    Pull("oldTags", "oldTag")

// Upsert 操作
bd.SetGetIdFunc(func() (any, bool) {
    return user.ID, true
})
_, err = up.UpsertOne(db.Sess(context.Background(), "mydb", "users"), bd)
```

### 聚合操作

```go
import "tm_orm/aggregator"

// 创建聚合器
agg := aggregator.NewAggregator[User]()

// 构建聚合管道
agg.Pipe().
    Match(func(m *query.MatchCmd) query.Builder {
        return m.K("name").Eq("sean")
    }).
    Group(func(gc *query.GroupCmd) query.Builder {
        return gc.IdWithField("name").
            Key("avgAge").Avg(gc.ToFd("age")).
            Key("count").Sum(gc.AnyVal(1)).Build()
    })

// 定义结果结构
type Result struct {
    Id     string  `bson:"_id"`
    AvgAge float64 `bson:"avgAge"`
    Count  int     `bson:"count"`
}

// 执行聚合并解析结果
var results []*Result
err := aggregator.WithParseAggregate(
    db.Sess(context.Background(), "mydb", "users"),
    agg,
    &results,
)
```
### 复杂聚合
```go
// 多级分组和条件聚合
agg.Pipe().
    Match(func(m *query.MatchCmd) query.Builder {
        return m.K("age").Gte(20)
    }).
    Group(func(gc *query.GroupCmd) query.Builder {
        return gc.Id(
            gc.IdBuilder().
                SetKeyField("department", "department").
                Key("ageRange").Cond(
                    gc.AggC().Lte(gc.ToFd("age"), gc.AnyVal(25)),
                    query.V("young"),
                    query.V("old"),
                ),
        ).
            Key("avgAge").Avg(gc.ToFd("age")).
            Key("names").Push(gc.ToFd("name")).Build()
    }).
    Sort("avgAge")
```

### 中间件使用

```go
import "tm_orm/middleware"

// 创建全局中间件调用链
mdb := tmorm.NewMDB(MongoClient)
mdb.GetMiddleware().Use(
	middleware.SLowQueryMiddleware{Threshold: 1}.Build(),
	func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
		return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
			println("db 全局前置")
			return next(mctx)
		}
	})

// 为单次查询session添加中间件
sess := md.Sess(context.Background(), ndb, ncoll,
		    func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
		    	return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		    		r := next(mctx)
		    		println("session 后置")
		    		return r
		    	}
		    },
        )
        
finder.Find(sess, nil)
```

## 支持的查询操作符

框架支持丰富的MongoDB查询操作符，包括：

- 比较操作符：`$eq`, `$gt`, `$gte`, `$lt`, `$lte`, `$ne`, `$in`, `$nin`
- 逻辑操作符：`$and`, `$or`
- 元素操作符：`$exists`, `$type`
- 评估操作符：`$expr`, `$regex`, `$mod`
- 数组操作符：`$all`, `$elemMatch`, `$size`

更多操作符支持详见[文档](doc/support_exprssion.md)。

## 许可证

本项目采用MIT许可证。详见[LICENSE](LICENSE)文件。
