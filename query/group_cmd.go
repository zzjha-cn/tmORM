package query

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
)

type (
	MatchCmd struct {
		Builder
	}

	GroupCmd struct {
		gb groupCmdBuilder
	}

	groupCmdBuilder struct {
		cur EBonsE
		agg aggCommand
		bd  Builder
	}
)

func NewGroupCmd() *GroupCmd {
	res := &GroupCmd{}
	res.gb.agg = newAggCommand()
	return res
}

// _id 可以是nil(null)，字段名，表达式，多级对象
func (g GroupCmd) IdWithField(fd string) GroupCmd {
	return g.id(F(fd))
}

func (g GroupCmd) Id(id IAggCommandField) GroupCmd {
	return g.id(id)
}

func (g *GroupCmd) AggC() aggCommand {
	return newAggCommand()
}

func (g *GroupCmd) IdBuilder() groupIdBuilder {
	return groupIdBuilder{
		agg: newAggCommand(),
	}
}

func (g *GroupCmd) ToFd(fd string) Field {
	return F(fd)
}

func (g *GroupCmd) AnyVal(v any) AggVal[any] {
	return V(v)
}

func (g GroupCmd) id(id IAggCommandField) GroupCmd {
	if id == nil {
		g.gb.bd = g.gb.bd.KV(tmorm.IdOp, nil)

	} else {
		g.gb.bd = g.gb.bd.KV(tmorm.IdOp, id.GetValue())
	}
	return g
}

func (g GroupCmd) Key(k string) groupCmdBuilder {
	g.gb.cur.SetKey(k)
	return g.gb
}

func (g GroupCmd) Raw(v bson.D) GroupCmd {
	g.gb.bd.data = append(g.gb.bd.data, v...)
	return g
}

func (g GroupCmd) Build() Builder {
	return g.gb.bd
}

// ===================================================================

func (gb groupCmdBuilder) Sum(v ...IAggCommandField) GroupCmd {
	builder := gb.agg.Sum(v...) // bson.D{{"$sum" , xx}}
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) toGroupCmd(bd Builder) GroupCmd {
	res := NewGroupCmd()
	res.gb.bd = gb.bd.KV(gb.cur.E.Key, bd.GetData())
	return *res
}

func (gb groupCmdBuilder) Avg(v ...IAggCommandField) GroupCmd {
	builder := gb.agg.Avg(v...)
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) AddToSet(v IAggCommandField) GroupCmd {
	builder := gb.agg.AddToSet(v)
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) First(v IAggCommandField) GroupCmd {
	builder := gb.agg.First(v)
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) Last(v IAggCommandField) GroupCmd {
	builder := gb.agg.Last(v)
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) Max(v ...IAggCommandField) GroupCmd {
	builder := gb.agg.Max(v...)
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) Min(v ...IAggCommandField) GroupCmd {
	builder := gb.agg.Min(v...)
	return gb.toGroupCmd(builder)
}

func (gb groupCmdBuilder) Push(v IAggCommandField) GroupCmd {
	builder := gb.agg.Push(v)
	return gb.toGroupCmd(builder)
}
