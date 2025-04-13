package aggregator

import (
	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/query"

	"go.mongodb.org/mongo-driver/bson"
)

func NewPipeline() *Pipeline {
	res := &Pipeline{}
	res.group = query.NewGroupCmd()
	return res
}

func (p *Pipeline) Match(f func(m *query.MatchCmd) query.Builder) *Pipeline {
	bd := f(&p.mt)
	p.mongoPl = append(p.mongoPl, bson.D{{tmorm.MatchOp, bd.GetData()}})
	return p
}

func (p *Pipeline) Group(f func(group *query.GroupCmd) query.Builder) *Pipeline {
	bd := f(p.group)
	p.mongoPl = append(p.mongoPl, bson.D{{tmorm.GroupOp, bd.GetData()}})
	return p
}

func (p *Pipeline) Sort(keys ...string) *Pipeline {
	var res bson.D = make(bson.D, 0, len(keys))
	for _, k := range keys {
		res = append(res, bson.E{k, 1})
	}
	p.mongoPl = append(p.mongoPl, bson.D{{tmorm.SortOp, res}})
	return p
}

func (p *Pipeline) Project(omitID bool, keys ...string) *Pipeline {
	var res bson.D = make(bson.D, 0, len(keys)+1)

	if !omitID {
		res = append(res, bson.E{"_id", 1})
	} else {
		res = append(res, bson.E{"_id", 0})
	}

	for _, k := range keys {
		res = append(res, bson.E{k, 1})
	}

	p.mongoPl = append(p.mongoPl, bson.D{{tmorm.ProjectOp, res}})
	return p
}

func (p *Pipeline) AppendRaw(v bson.D) *Pipeline {
	p.mongoPl = append(p.mongoPl, v)
	return p
}
