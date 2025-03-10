package query

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
)

type (
	groupIdBuilder struct {
		cur EBonsE
		agg aggCommand
		bd  Builder
	}
)

func (gb groupIdBuilder) cmdFd() {}
func (gb groupIdBuilder) GetValue() any {
	return gb.bd.GetValue()
}

func (gb groupIdBuilder) SetKeyField(name string, fd string) groupIdBuilder {
	gb.bd = gb.bd.KV(name, F(fd).GetValue())
	return gb
}

func (gb groupIdBuilder) Key(k string) groupIdBuilder {
	gb.cur.SetKey(k)
	return gb
}

func (gb groupIdBuilder) SetValue(v IAggCommandField) groupIdBuilder {
	res := groupIdBuilder{}
	res.agg = newAggCommand()
	res.bd = gb.bd.KV(gb.cur.E.Key, v.GetValue())
	return res
}

// =============================================================

func (gb groupIdBuilder) Year(v IAggCommandField) groupIdBuilder {
	builder := gb.agg.combineSingleVal(tmorm.YearOp, v)
	return gb.merge(builder)
}

func (gb groupIdBuilder) merge(bd Builder) groupIdBuilder {
	res := groupIdBuilder{}
	res.agg = newAggCommand()
	if len(gb.cur.E.Key) > 0 {
		// {_id: {
		//    y1 : { $year: "$cur_date"},
		//    m1: {$month: "$cur_date"}
		// }}
		res.bd = gb.bd.KV(gb.cur.E.Key, bd.GetData())
	} else {
		// {_id: {$year: "$cur_date"}}
		res.bd.data = append(gb.bd.data, bd.GetData()...)
	}
	return res
}

func (gb groupIdBuilder) Month(v IAggCommandField) groupIdBuilder {
	builder := gb.agg.combineSingleVal(tmorm.MonthOp, v)
	return gb.merge(builder)
}

func (gb groupIdBuilder) DateToString(v IAggCommandField, format, timezone, onNull any) groupIdBuilder {
	data := bson.D{
		{"date", v.GetValue()},
	}
	if format != nil {
		data = append(data, bson.E{"format", format})
	}
	if timezone != nil {
		data = append(data, bson.E{"timezone", timezone})
	}
	if onNull != nil {
		data = append(data, bson.E{"onNull", onNull})
	}

	builder := gb.agg.combineSingleVal(tmorm.DateToStringOp, V(data))
	return gb.merge(builder)
}

func (gb groupIdBuilder) ArrayElemAt(k IAggCommandField, v IAggCommandField) groupIdBuilder {
	builder := gb.agg.ArrayElemAt(k, v)
	return gb.merge(builder)
}

func (gb groupIdBuilder) Type(k IAggCommandField) groupIdBuilder {
	builder := gb.agg.Type(k)
	return gb.merge(builder)
}

func (gb groupIdBuilder) Cond(ifExpr, thenVal, elseVal IAggCommandField) groupIdBuilder {
	builder := gb.agg.Cond(ifExpr, thenVal, elseVal)
	return gb.merge(builder)
}
