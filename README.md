# tmORM

tmORM 是一个用 Go 语言编写的 MongoDB ORM 框架，提供了简洁优雅的 API 接口，让 MongoDB 的数据库操作更加便捷。

## 特性

- 链式查询构建器，支持复杂查询条件的优雅构建
- 支持丰富的查询操作符(in, nin, gt, gte, lt, lte, eq, ne, exists, regex等)
- 支持逻辑组合查询(and, or)
- 支持表达式查询($expr)
- 支持字段映射和自定义标签
- 支持 CRUD 操作
- 支持 Upsert 操作

## 安装

```bash
go get 
```

## 快速开始

### 连接数据库

```go
import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    tmorm "tm_orm"
)

// 创建MongoDB客户端
client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    panic(err)
}

// 初始化tmORM
db := tmorm.NewMDB(client)
```

### 定义模型

```go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Age       int64              `bson:"age"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

### 查询示例

```go
// 创建会话
sess := db.Sess(context.Background(), "mydb", "users")

// 构建查询条件
query := query.Query{}.Builder().
    K("name").Eq("sean").
    K("age").Gte(18).ToQuery()

// 执行查询
finder := &finder.Finder[User]{}
users, err := finder.Find(sess, query)
```

### 复杂查询示例

```go
// AND 查询
query := query.Query{}.Builder().
    And(func(a *query.QueryAnd) query.Builder {
        return a.K("name").Eq("sean").
            K("age").Gt(20)
    }).ToQuery()

// OR 查询
query := query.Query{}.Builder().
    Or(func(a *query.QueryOr) query.Builder {
        return a.K("age").Gt(10).
            K("name").Eq("sean")
    }).ToQuery()

// 表达式查询
query := query.Query{}.Builder().
    Expr(func(m query.MExpr) query.Builder {
        return m.Gt(
            m.C().Multi(m.Fd("age"), 2),
            40,
        )
    }).ToQuery()
```

### 更新示例

```go
// 创建更新器
updater := &updater.MUpdater[User]{}

// 构建更新条件
query := query.Query{}.Builder().
    K("name").Eq("sean").ToQuery()

// 构建更新内容
updateBuilder := updater.NewUpdateBuilder(&User{
    Name: "new_name",
    Age:  25,
})
updateBuilder.OmitZero = true

// 执行更新
result, err := updater.Filter(query).UpdateOne(sess, updateBuilder)
```

## 许可证

本项目采用 MIT 许可证，详情请参见 LICENSE 文件。
