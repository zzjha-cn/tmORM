package test

import (
	"context"
	"fmt"
	"testing"
	"time"
	tmorm "tm_orm"
	"tm_orm/finder"
	"tm_orm/impl"
	"tm_orm/middleware"
	"tm_orm/query"
	"tm_orm/updater"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	dbName = ""
	coll   = ""
)

func getDB() *tmorm.MDB {
	return &tmorm.MDB{}
}

func TestUserType(t *testing.T) {
	// 常规使用

	// - 获取
	ctx := context.Background()
	sess := getDB().Sess(ctx, dbName, coll)

	// - 查找
	q := query.Query{}
	u, _ := (&finder.Finder[TestUser]{}).Find(sess, q)
	fmt.Println(u)

	// - 构造
	// bsonD{{"",bsonD{{"$gte",1}}}}
	q1 := query.Query{}
	q1.Builder().KV("", TestUser{})
	q1.Builder().K("").Gte(1)
	q1.Builder().K("").In(1, 2, 3, 4)

	// - and
	// bsonD{{"$and", bsonD{{"age" , bsonD{{"$gte",1}} } ,  }}}
	q1.Builder().And(
		func(a *query.QueryAnd) query.Builder {
			return a.K("name").Gt("").
				K("age").Eq(324)
		},
	)
	q1.Builder().Or(
		func(a *query.QueryOr) query.Builder {
			return a.K("name").Gt("").
				K("age").Eq(324)
		},
	)
	//$and: [
	//  { $expr: { $gt: ["$age", 30] } },  // 年龄大于30岁
	//  { $expr: { $gt: ["$salary", 10000] } }  // 薪资高于10000
	//]
	q1.Builder().And(
		func(a *query.QueryAnd) query.Builder {
			return a.
				Expr(func(m query.MExpr) query.Builder {
					return m.Or(m.C().Eq(m.Fd("name"), "sean2"))
				}).
				Expr(func(m query.MExpr) query.Builder {
					return m.Gt(m.Fd("salary"), 1000)
				})
		},
	)

	//- expr
	// bson.D{{ "$expr" , bson.D{{ "$gte" , bson.A{"$age", 33} }} }}
	//q1.Builder().Expr().AggCmd().Gte(query.F("age"), 33)
	q1.Builder().Expr(func(m query.MExpr) query.Builder {
		return m.Lte(m.Fd("age"), 33)
	})
	q1.Builder().Expr(func(m query.MExpr) query.Builder {
		return m.Gt(
			m.C().Multi(m.Fd("age"), 10),
			50,
		)
	})
	q1.Builder().Expr(func(m query.MExpr) query.Builder {
		return m.And(
			m.C().Eq(m.Fd("name"), "sean"),
			m.Or(),
		)
	})
	fmt.Println(u)

	// - update
	up := &updater.MUpdater[TestUser]{}
	q2 := updater.NewReplaceBuilder[TestUser]()
	q2.SetGetIdFunc(func() (any, bool) {
		return 1, true
	})
	q2.C().SetObj(&TestUser{Name: "s1"}, false)
	_, _ = up.SetFilter(q1).UpdateOne(sess, q2)
	_, _ = up.SetFilter(q1).UpdateMany(sess, q2)

	// - update replace
	_, _ = up.SetFilter(q1).ReplaceOne(sess, updater.NewBaseSetBuilder(&struct {
		Name  string `bson:"name"`
		Age   int    `bson:"age"`
		Score int    `bson:"-"`
	}{}))

	// - update set
	q2.C().SetObj(&TestUser{}, true)
	q2.C().SetObj(&TestUser{}, false)
	q2.C().Set("", "").Unset("", "").Rename("nn", "n2")
	q2.C().Set("", "").AddToSet("letters", []int{1, 2, 3}).Min("m", 3)
	up.
		CommonFilter(func(q query.Query) impl.IBsonQuery {
			return q.Builder().K("age").Lte(13).ToQuery()
		}).
		UpsertOne(sess, q2)

	// - middleware
	md := tmorm.NewMiddleChainAdapt()
	md.Use( // DB层面的中间件
		func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
			return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
				println("前置")
				return next(mctx)
			}
		},
		func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
			return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
				r := next(mctx)
				println("后置")
				return r
			}
		},
		middleware.SLowQueryMiddleware{
			Threshold: 500, // ms
		}.Build(),
	)
	ctx1 := context.Background()
	db := getDB().SetMiddleware(md)
	sess1 := db.Sess(ctx1, dbName, coll, func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
		return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
			println("sess 前置，会话层面的中间件")
			return next(mctx)
		}
	})
	(&finder.Finder[TestUser]{}).Find(sess1, q1)

}

type (
	TestUser struct {
		ID           primitive.ObjectID `bson:"_id,omitempty"`
		Name         string             `bson:"name"`
		Age          int64              `bson:"age"`
		UnknownField string             `bson:"-"`
		CreatedAt    time.Time          `bson:"created_at"`
		UpdatedAt    time.Time          `bson:"updated_at"`
	}

	Tquery struct {
		E bson.E
	}

	mongoCfg struct {
		Username      string
		Password      string
		AuthMechanism string
		AuthSource    string

		Timeout        int
		HostsWithPorts []string
		MaxPool        uint64
		MinPool        uint64
		ReplicaSet     string
	}
)

func (t *Tquery) GetBsonD() bson.D {
	println(t.E.Value)
	a := bson.D{}
	a = append(a, t.E)
	return a
}
