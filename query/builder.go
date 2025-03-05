package query

import (
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	IBsonQuery interface {
		GetBsonD() bson.D
	}

	IBsonBuilder interface {
	}
)

type (
	Builder struct {
		data bson.D
	}
)

func (q Builder) CmdKey() {}

// ====================================================

func (q Builder) KV(field string, v any) Builder {
	q.data = appendBsonD(q.data, field, v)
	return q
}

func appendBsonD(D bson.D, key string, val any) bson.D {
	D = append(D, bson.E{Key: key, Value: val})
	return D
}

func (q Builder) K(field string) mongoCmd {
	mc := newMCommand()
	mc.e.SetKey(field)
	mc.b = &q
	return mc
}

func (q Builder) Expr(f func(m MExpr) Builder) Builder {
	mc := newAggCommand()
	mc.m.b = &Builder{}
	ex := MExpr{mc}

	exprBd := f(ex)

	return q.KV(tmorm.ExprOp, exprBd.data)
}

func (q Builder) And(f func(a *QueryAnd) Builder) Builder {
	qa := newQueryAnd()
	mc := newMCommand()
	qa.mc = mc
	qa.mc.b = &Builder{}

	andBd := f(qa)

	arr := bson.A{}
	for _, e := range andBd.data {
		arr = append(arr, bson.D{e})
	}

	return q.KV(tmorm.AndOp, arr)
}

func (q Builder) Or(f func(a *QueryOr) Builder) Builder {
	qa := newQueryOr()
	mc := newMCommand()
	qa.mc = mc
	qa.mc.b = &Builder{}

	andBd := f(qa)

	arr := bson.A{}
	for _, e := range andBd.data {
		arr = append(arr, bson.D{e})
	}

	return q.KV(tmorm.OrOp, arr)
}

func (q Builder) ToQuery() Query {
	return Query{
		bd: q,
	}
}

func (q Builder) GetData() bson.D {
	return q.data
}
