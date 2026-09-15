<p align="right">
   <strong>English</strong> | <a href="./README_CN.md">中文</a>
</p>

# tmORM

tmORM is a lightweight MongoDB ORM for Go. It gives you a type-safe, chainable query builder on top of the official
`mongo-driver`, plus update builders, an aggregation pipeline builder, and a middleware chain for cross-cutting
concerns such as logging, retries and circuit breaking.

All examples in this README are compiled against the current API.

## Features

- Type-safe, generic query builder (`Collection[T]`)
- Composable query conditions: chained `Where`, expression trees, raw BSON
- Update builder covering `$set` / `$inc` / `$mul` / array operators / `$rename` / `$currentDate`
- Aggregation pipeline builder with typed results
- Repository pattern with common CRUD helpers
- Middleware chain with built-in logging, recovery, timing, retry, error-handling and circuit-breaker
- Structured results with error classification (`MResult`)
- Multi-client management and transaction helpers

## Requirements

- Go 1.21+
- MongoDB 4.2+ (transactions require a replica set)

## Installation

```bash
go get github.com/zzjha-cn/tm_orm
```

## Quick Start

### Connect

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

`NewORMClient` connects and pings eagerly, so a non-nil return value is already usable. Passing `nil` as the config
falls back to `mongodb://localhost:27017`.

### Define a model and query it

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

// Find many
users, err := coll.Where("name").Eq("sean").Where("age").Gte(18).Find(ctx)

// Find one
user, err := coll.Where("_id").Eq(userID).FindOne(ctx)

// Count
count, err := coll.Where("age").Gte(18).Count(ctx)

// Native driver options are passed straight through
opts := options.Find().SetSort(bson.D{{Key: "age", Value: -1}}).SetLimit(10)
users, err = coll.Where("age").Gte(18).Find(ctx, opts)
```

## Queries

### Chained conditions

Every `Where(...)` in a chain is combined with `$and`, so repeating the same field works as expected:

```go
users, err := coll.
    Where("age").Gte(18).Where("age").Lte(65).
    Where("name").Regex("^[A-Z]").
    Where("email").Exists(true).
    Find(ctx)
```

produces

```js
{ $and: [ { age: { $gte: 18 } }, { age: { $lte: 65 } },
          { name: { $regex: "^[A-Z]" } }, { email: { $exists: true } } ] }
```

A `Collection` is immutable: each condition returns a new value, so a filtered collection can be safely reused as a
base for several queries.

Available on `Where(field)`: `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `In`, `NotIn`, `Regex`, `Exists`, and `Not()` to
negate the next operator.

```go
// Membership
users, err := coll.Where("city").In("Beijing", "Shanghai", "Guangzhou").Find(ctx)

// Negation -> { status: { $not: { $eq: "active" } } }
users, err = coll.Where("status").Not().Eq("active").Find(ctx)
```

### Raw BSON filters

`Filter` and `AndFilter` accept anything implementing `impl.IBsonQuery`, which is a single method:

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

// Filter REPLACES any conditions set so far
users, err := coll.Filter(query).Find(ctx)

// AndFilter APPENDS to them with $and
users, err = coll.Where("age").Gte(18).AndFilter(query).Find(ctx)
```

### Expressions

`Expr(field)` exposes a richer operator set than `Where`, including string, range, array and date helpers:

```go
users, err := coll.Expr("age").Between(18, 65).Find(ctx)
users, err = coll.Expr("email").Contains("@company.com").Find(ctx)
users, err = coll.Expr("created_at").LastNDays(7).Find(ctx)
users, err = coll.Expr("tags").All("golang", "backend").Find(ctx)

// Each Expr(...) terminal returns a Collection, so they chain with $and
users, err = coll.Expr("age").Gt(25).Expr("status").Eq("active").Find(ctx)
```

`Expr(field)` supports `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `In`, `NotIn`, `Exists`, `Regex`, `Contains`,
`StartsWith`, `EndsWith`, `Between`, `IsNull`, `IsNotNull`, `Size`, `All`, `ElemMatch`, and the date shortcuts
`Today`, `Yesterday`, `ThisWeek`, `ThisMonth`, `LastNDays`.

For boolean logic, build expressions with the `expression` package and combine them with `And` / `Or` / `Nor` / `Not`:

```go
import "github.com/zzjha-cn/tm_orm/expression"

expr1 := expression.NewExpression().Field("name").Eq("sean")
expr2 := expression.NewExpression().Field("age").Gt(25)

users, err := coll.WhereExpr(expression.Or(expr1, expr2)).Find(ctx)
users, err = coll.WhereExpr(expression.And(expr1, expr2)).Find(ctx)
users, err = coll.WhereExpr(expression.Not(expr1)).Find(ctx)

// expression.Q is a shorthand for NewExpression().Field(...)
nested := expression.And(
    expression.Or(
        expression.Q("department").Eq("Engineering"),
        expression.Q("department").Eq("IT"),
    ),
    expression.Q("age").Between(25, 45),
)
users, err = coll.WhereExpr(nested).Find(ctx)
```

> **Note:** a single `Expression` carries one logical operator. Calling `.And()` / `.Or()` on it switches that
> operator for *all* conditions it holds — it does not mix them. To combine AND and OR, nest `expression.And(...)`
> and `expression.Or(...)` as shown above. `WhereExpr` itself appends with `$and`, so it also composes with `Where`.

## Updates

```go
// Single field
res, err := coll.Where("_id").Eq(userID).Set("age", 25).UpdateOne(ctx)

// Multiple fields, all matching documents
res, err = coll.Where("name").Eq("sean").
    Set("age", 26).
    Set("email", "sean@example.com").
    Update(ctx)

// Numeric operators
res, err = coll.Where("_id").Eq(userID).
    Inc("age", 1).
    Mul("salary", 1.1).
    UpdateOne(ctx)

// Field removal / rename / timestamps
res, err = coll.Where("_id").Eq(userID).
    Unset("deprecated_field").
    Rename("old_field", "new_field").
    CurrentDate("updated_at").
    UpdateOne(ctx)

// Upsert
res, err = coll.Where("name").Eq("sean").
    Set("age", 30).
    SetOnInsert("created_at", time.Now()).
    UpdateOne(ctx, options.Update().SetUpsert(true))
```

`Collection` starts an update chain via `Set`, `Inc` or `Unset`. Every other operator — including the array ones —
lives on `UpdateBuilder`, so start from one of those three or construct the builder explicitly:

```go
res, err := collection.NewUpdateBuilder(coll.Where("_id").Eq(userID)).
    Push("tags", "new-tag").
    Pull("old_tags", "deprecated-tag").
    AddToSet("skills", "golang").
    UpdateOne(ctx)
```

`UpdateOne` maps to `UpdateOne`, `Update` maps to `UpdateMany`.

> **Careful:** `Update` and `Delete` without any condition match the entire collection. Always set a filter unless
> that is really what you want.

## Insert and Delete

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

list, total, err := r.FindPage(ctx, 1, 10)                    // page 1, 10 per page
exists, err := r.Exists(ctx, id)
count, err := r.Count(ctx)
```

`BaseRepository` adds field-oriented helpers on top:

```go
br := repo.NewBaseRepository[User](ormClient, "mydb", "users")

byField, err := br.FindByField(ctx, "name", "sean")
one, err := br.FindOneByField(ctx, "email", "sean@example.com")
in, err := br.FindIn(ctx, "city", "Beijing", "Shanghai")
n, err := br.CountByField(ctx, "status", "active")
```

Note that `FindPage` counts the whole collection, not the filtered subset.

## Aggregation

`Aggregator[T]` decodes results into `T`, so parameterise it with the *shape of the output*, not the source
document:

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

`$match` can also be built from an expression:

```go
stats, err := agg.
    MatchByExpression(expression.Q("status").Eq("active")).
    Group("$department", bson.M{"count": bson.M{"$sum": 1}}).
    Execute(ctx)
```

### Joins

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
    Unwind("user_info").                 // the leading "$" is added for you
    Project(bson.M{
        "_id":        1,
        "amount":     1,
        "user_name":  "$user_info.name",
        "user_email": "$user_info.email",
    }).
    Execute(ctx)

// Keep documents whose array is empty or missing
orders, err = orderAgg.Unwind("user_info", true).Execute(ctx)
```

### Other stages

```go
// Facet
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

// Pagination: Paginate only appends $skip/$limit, you still call Execute
page, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "users").
    Match(bson.M{"age": bson.M{"$gte": 18}}).
    SortAsc("name").
    Paginate(1, 10).
    Execute(ctx)

// $replaceRoot takes { newRoot: <expr> }; $replaceWith takes the expression directly
rooted, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "orders").
    Unwind("items").
    ReplaceRoot("$items").
    Execute(ctx)

withed, err := aggregator.NewAggregator[bson.M](ormClient, "mydb", "orders").
    ReplaceWith("$items").
    Execute(ctx)
```

`ExecuteOne` returns the first document (`mongo.ErrNoDocuments` when empty) and `ExecuteRaw` hands back the native
cursor:

```go
single, err := agg.Match(bson.M{"name": "sean"}).ExecuteOne(ctx)

cursor, err := agg.Match(bson.M{"name": "sean"}).ExecuteRaw(ctx)
if err == nil {
    defer cursor.Close(ctx)
}
```

> **Note:** unlike `Collection`, an `Aggregator` is mutable — every stage is appended to the same pipeline. Create a
> new aggregator per query rather than reusing one, and be aware that `ExecuteOne` appends a `$limit` stage that
> stays in the pipeline.

## Middleware

A middleware is a function that receives the operation context and the rest of the chain. Client-level middleware
runs first, then collection-level middleware, then the actual database call.

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

// Client level: applies to every operation on this client
ormClient.Use(
    tmorm.Recovery(),
    tmorm.LoggingMiddleware(log.Default()),
    tmorm.TimingMiddleware(100*time.Millisecond),
    tmorm.RetryMiddleware(3, 200*time.Millisecond),
    tmorm.CircuitBreakerMiddleware(5, 30*time.Second),
)

// Collection level: applies to this collection only
coll := collection.NewCollection[User](ormClient, "mydb", "users", loggingMiddleware)
```

Built-ins:

| Middleware | Purpose |
| --- | --- |
| `Recovery()` | Recovers panics into a fatal `MResult` with a stack trace |
| `LoggingMiddleware(logger)` | Logs operation, filter, duration and error details |
| `TimingMiddleware(threshold)` | Flags slow operations as a warning |
| `RetryMiddleware(max, delay)` | Retries retryable errors (network / timeout) |
| `ErrorHandlingMiddleware(cb)` | Invokes a callback with the classified error |
| `CircuitBreakerMiddleware(n, reset)` | Opens after `n` consecutive failures |

Inside a middleware, `mctx.Operation` is the filter, `mctx.UpdateOp` the update document, `mctx.AggOp` the pipeline,
and `mctx.Result` the outcome.

### Errors

Every operation fills an `MResult`. Errors are classified automatically into a type, a code and a retryable flag:

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

Error types: `CONNECTION`, `TIMEOUT`, `VALIDATION`, `QUERY`, `UPDATE`, `INSERT`, `DELETE`, `AGGREGATE`, `AUTH`,
`INDEX`, `UNKNOWN`. Levels: `INFO`, `WARNING`, `ERROR`, `FATAL`.

## Transactions

Transactions require a replica set or a sharded cluster. Pass the session context to every operation inside:

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

## Multiple clients

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

A process-wide manager is also available via `tmorm.GlobalManager()`, `tmorm.AddGlobalClient(...)` and
`tmorm.GetGlobalClient(...)`.

## Supported operators

- Comparison: `$eq`, `$gt`, `$gte`, `$lt`, `$lte`, `$ne`, `$in`, `$nin`
- Logical: `$and`, `$or`, `$not`, `$nor`
- Element: `$exists`, `$type`
- Evaluation: `$expr`, `$regex`, `$mod`, `$text`
- Array: `$all`, `$elemMatch`, `$size`
- Update: `$set`, `$unset`, `$inc`, `$mul`, `$rename`, `$setOnInsert`, `$currentDate`, `$push`, `$pop`, `$pull`,
  `$pullAll`, `$addToSet`
- Aggregation stages: `$match`, `$group`, `$sort`, `$project`, `$limit`, `$skip`, `$unwind`, `$lookup`,
  `$graphLookup`, `$addFields`, `$replaceRoot`, `$replaceWith`, `$facet`, `$bucket`, `$sample`, `$out`, `$merge`,
  `$redact`, `$geoNear`, `$count`

See [docs/support_exprssion.md](docs/support_exprssion.md) for the full list.

## Caveats

Things worth knowing before you hit them:

- A single `Expression` holds one logical operator; mix AND and OR by nesting `expression.And` / `expression.Or`
  rather than chaining `.And()` / `.Or()` on the same expression.
- `Aggregator` mutates its pipeline in place — use a fresh one per query.
- `Aggregator.PaginateWithCount` is not implemented yet; use `Paginate` plus a separate count.
- `GroupBuilder` and the `MatchOperators` / `ProjectOperators` / `SortOperators` helpers `Build()` into
  `Aggregator[any]`, which loses the result type. Prefer `Group(id, fields)` and `Match(filter)` for typed results.
- `Update` and `Delete` with no filter affect the whole collection.
- `Repository.FindPage` returns the total for the entire collection, not for the filter.
- Middleware cannot currently override the `context.Context` used by the driver call.

## Roadmap

- [x] Query builder
- [x] TDD test suite
- [x] Query expressions, AND / OR
- [x] Aggregation operators
- [x] Middleware
- [x] Aggregator (basic)
- [x] End-to-end tests
- [x] Error classification and error callbacks
- [ ] Full coverage of aggregation stage helpers
- [ ] Metadata management
- [ ] Atomic operation helpers
- [ ] Result-set processing helpers

## Benchmarks

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

## License

MIT. See [LICENSE](LICENSE).
