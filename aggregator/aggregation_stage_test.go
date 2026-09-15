package aggregator

import (
	"testing"

	tmorm "github.com/zzjha-cn/tm_orm"
	"go.mongodb.org/mongo-driver/bson"
)

type aggTestDoc struct {
	Name string `bson:"name"`
}

// $replaceRoot 需要 {newRoot: <expr>}，不能写成 $replaceWith
func TestReplaceRootStageShape(t *testing.T) {
	a := NewAggregator[aggTestDoc](nil, "testdb", "testcoll").ReplaceRoot("$sub")

	pipeline := a.Pipeline()
	if len(pipeline) != 1 {
		t.Fatalf("期望 1 个阶段，实际 %d 个", len(pipeline))
	}

	stage := pipeline[0]
	if _, ok := stage[tmorm.ReplaceWithOp]; ok {
		t.Fatalf("ReplaceRoot 不应生成 %s: %v", tmorm.ReplaceWithOp, stage)
	}

	spec, ok := stage[tmorm.ReplaceRootOp].(bson.M)
	if !ok {
		t.Fatalf("期望 %s 阶段，实际 %v", tmorm.ReplaceRootOp, stage)
	}
	if spec["newRoot"] != "$sub" {
		t.Fatalf("newRoot 不匹配: %v", spec)
	}
}

// $replaceWith 直接接受表达式，没有 newRoot 包装
func TestReplaceWithStageShape(t *testing.T) {
	a := NewAggregator[aggTestDoc](nil, "testdb", "testcoll").ReplaceWith("$sub")

	stage := a.Pipeline()[0]
	if got := stage[tmorm.ReplaceWithOp]; got != "$sub" {
		t.Fatalf("期望 %s 直接接受表达式，实际 %v", tmorm.ReplaceWithOp, stage)
	}
}
