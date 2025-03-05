package updater

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
	"tm_orm/query"
)

type (
	UpdateCmd[T any] struct {
		*updateCmd
	}

	updateCmd struct {
		bd query.Builder
		fdCmd
		adCmd
	}

	fdCmd struct {
		source *updateCmd
	}

	adCmd struct {
		source *updateCmd
	}
)

func newUpdateCmd[T any]() *UpdateCmd[T] {
	q := &updateCmd{}
	q.fdCmd.source = q
	q.adCmd.source = q

	res := &UpdateCmd[T]{}
	res.updateCmd = q

	return res
}

func (c *UpdateCmd[T]) SetObj(t *T, omitZero ...bool) *UpdateCmd[T] {
	o := false
	if len(omitZero) > 0 {
		o = omitZero[0]
	}
	d, _ := makeBsonDByReflect(t, o)
	c.bd = c.bd.KV(tmorm.SetOp, d)
	return c
}

func (b fdCmd) Set(key string, value any) updateCmd {
	return b.appendKv(tmorm.SetOp, key, value)
}

func (b fdCmd) Unset(keys ...string) updateCmd {
	value := bson.D{}
	for i := range keys {
		value = append(value, bson.E{Key: keys[i], Value: ""})
	}
	b.source.bd = b.source.bd.KV(tmorm.UnsetOp, value)
	return *(b.source)
}

func (b fdCmd) appendKv(op, key string, value any) updateCmd {
	e := bson.E{Key: key, Value: value}
	//(*b.source).bd = b.source.bd.KV(op, bson.D{e})
	b.source.bd = b.source.bd.KV(op, bson.D{e})
	return *(b.source)
}

func (b fdCmd) SetOnInsert(key string, value any) updateCmd {
	return b.appendKv(tmorm.SetOnInsertOp, key, value)
}

func (b fdCmd) CurrentDate(key string, value any) updateCmd {
	return b.appendKv(tmorm.CurrentDateOp, key, value)

}

func (b fdCmd) Inc(key string, value any) updateCmd {
	return b.appendKv(tmorm.IncOp, key, value)

}

func (b fdCmd) Min(key string, value any) updateCmd {
	return b.appendKv(tmorm.MinOp, key, value)
}

func (b fdCmd) Max(key string, value any) updateCmd {
	return b.appendKv(tmorm.MaxOp, key, value)
}

func (b fdCmd) Mul(key string, value any) updateCmd {
	return b.appendKv(tmorm.MulOp, key, value)
}

func (b fdCmd) Rename(key string, value any) updateCmd {
	return b.appendKv(tmorm.RenameOp, key, value)
}

//==============================================================

func (b adCmd) appendKv(op, key string, value any) updateCmd {
	e := bson.E{Key: key, Value: value}
	b.source.bd = b.source.bd.KV(op, bson.D{e})
	return *b.source
}

func (b adCmd) AddToSet(key string, value any) updateCmd {
	return b.appendKv(tmorm.AddToSetOp, key, value)
}

func (b adCmd) Pop(key string, value int) updateCmd {
	return b.appendKv(tmorm.PopOp, key, value)
}

func (b adCmd) Pull(key string, value any) updateCmd {
	return b.appendKv(tmorm.PullOp, key, value)
}

func (b adCmd) Push(key string, value any) updateCmd {
	return b.appendKv(tmorm.PushOp, key, value)
}

func (b adCmd) PullAll(key string, values ...any) updateCmd {
	a := make(bson.A, 0, len(values))
	for _, v := range values {
		a = append(a, v)
	}
	return b.appendKv(tmorm.PullAllOp, key, a)
}

func (b adCmd) Each(values ...any) updateCmd {
	a := make(bson.A, 0, len(values))
	for _, v := range values {
		a = append(a, v)
	}
	b.source.bd = b.source.bd.KV(tmorm.EachOp, a)
	return *b.source
}

func (b adCmd) Position(val uint) updateCmd {
	b.source.bd = b.source.bd.KV(tmorm.PositionOp, val)
	return *b.source
}

func (b adCmd) Slice(num int) updateCmd {
	b.source.bd = b.source.bd.KV(tmorm.SliceForUpdateOp, num)
	return *b.source
}

func (b adCmd) Sort(value any) updateCmd {
	b.source.bd = b.source.bd.KV(tmorm.SortOp, value)
	return *b.source
}
