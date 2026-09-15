package collection

import (
	"testing"

	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/expression"
	"github.com/zzjha-cn/tm_orm/impl"

	"go.mongodb.org/mongo-driver/bson"
)

type filterTestDoc struct {
	Name string `bson:"name"`
	Age  int64  `bson:"age"`
}

// filterOf 取出集合当前的过滤条件，供断言使用
func filterOf[T any](c *Collection[T]) bson.D {
	if c.filter == nil {
		return nil
	}
	return c.filter.GetBsonD()
}

func newTestCollection() *Collection[filterTestDoc] {
	return NewCollection[filterTestDoc](nil, "testdb", "testcoll")
}

// 链式多条件必须全部保留，而不是只留最后一个
func TestChainedWhereKeepsAllConditions(t *testing.T) {
	c := newTestCollection().
		Where("age").Gte(int64(18)).
		Where("age").Lte(int64(65)).
		Where("name").Eq("sean")

	got := filterOf(c)
	if len(got) != 1 || got[0].Key != tmorm.AndOp {
		t.Fatalf("期望顶层为 %s，实际为 %v", tmorm.AndOp, got)
	}

	conds, ok := got[0].Value.(bson.A)
	if !ok {
		t.Fatalf("$and 的值应为 bson.A，实际为 %T", got[0].Value)
	}
	if len(conds) != 3 {
		t.Fatalf("期望保留 3 个条件，实际 %d 个: %v", len(conds), conds)
	}

	want := []bson.D{
		{{Key: "age", Value: bson.D{{Key: tmorm.GteOp, Value: int64(18)}}}},
		{{Key: "age", Value: bson.D{{Key: tmorm.LteOp, Value: int64(65)}}}},
		{{Key: "name", Value: bson.D{{Key: tmorm.EqOp, Value: "sean"}}}},
	}
	for i, w := range want {
		if got := conds[i]; !equalBsonD(got.(bson.D), w) {
			t.Errorf("第 %d 个条件不匹配\n got: %v\nwant: %v", i, got, w)
		}
	}
}

// 单个条件不应被包进多余的 $and
func TestSingleConditionNotWrapped(t *testing.T) {
	c := newTestCollection().Where("age").Gt(int64(10))
	got := filterOf(c)
	want := bson.D{{Key: "age", Value: bson.D{{Key: tmorm.GtOp, Value: int64(10)}}}}
	if !equalBsonD(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// 派生出的集合之间不能相互污染
func TestChainedWhereDoesNotMutateSource(t *testing.T) {
	base := newTestCollection().Where("age").Gt(int64(10))
	baseBefore := filterOf(base)

	_ = base.Where("name").Eq("a")
	_ = base.Where("name").Eq("b")

	if !equalBsonD(filterOf(base), baseBefore) {
		t.Fatalf("派生查询污染了源集合: %v -> %v", baseBefore, filterOf(base))
	}
}

// 多次 $and 追加需要展平，不能层层嵌套
func TestAndFilterFlattens(t *testing.T) {
	c := newTestCollection().
		Where("a").Eq(1).
		Where("b").Eq(2).
		Where("c").Eq(3)

	got := filterOf(c)
	conds := got[0].Value.(bson.A)
	for i, cond := range conds {
		d := cond.(bson.D)
		if len(d) == 1 && d[0].Key == tmorm.AndOp {
			t.Fatalf("第 %d 个条件出现了嵌套的 $and: %v", i, d)
		}
	}
}

// Filter 保持覆盖语义
func TestFilterReplaces(t *testing.T) {
	c := newTestCollection().Where("a").Eq(1)
	replaced := c.Filter(&SimpleQuery{data: bson.D{{Key: "b", Value: 2}}})

	got := filterOf(replaced)
	want := bson.D{{Key: "b", Value: 2}}
	if !equalBsonD(got, want) {
		t.Fatalf("Filter 应覆盖已有条件，got %v want %v", got, want)
	}
}

// Where 与 WhereExpr 混用时同样要累加
func TestWhereExprAppends(t *testing.T) {
	c := newTestCollection().
		Where("age").Gt(int64(10)).
		WhereExpr(expression.NewExpression().Field("name").Eq("sean"))

	got := filterOf(c)
	if len(got) != 1 || got[0].Key != tmorm.AndOp {
		t.Fatalf("期望顶层为 %s，实际为 %v", tmorm.AndOp, got)
	}
	if conds := got[0].Value.(bson.A); len(conds) != 2 {
		t.Fatalf("期望 2 个条件，实际 %d 个: %v", len(conds), conds)
	}
}

// 空条件不应产生无意义的 $and 项
func TestAndQuerySkipsEmptyParts(t *testing.T) {
	q := &AndQuery{parts: []impl.IBsonQuery{
		&SimpleQuery{data: bson.D{}},
		&SimpleQuery{data: bson.D{{Key: "a", Value: 1}}},
		nil,
	}}
	want := bson.D{{Key: "a", Value: 1}}
	if got := q.GetBsonD(); !equalBsonD(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got := (&AndQuery{}).GetBsonD(); len(got) != 0 {
		t.Fatalf("空 AndQuery 应返回空条件，实际 %v", got)
	}
}

func equalBsonD(a, b bson.D) bool {
	ab, err1 := bson.Marshal(a)
	bb, err2 := bson.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(ab) == string(bb)
}
