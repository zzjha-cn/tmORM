package aggregator

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
	"tm_orm/query"
)

func (p *Pipeline) Match(f func(m *query.MatchCmd) query.Builder) *Pipeline {
	bd := f(&p.mt)
	p.pl = append(p.pl, bson.D{{tmorm.MatchOp, bd}})
	return p
}

func (p *Pipeline) Group(f func(group *query.GroupCmd) query.Builder) *Pipeline {
	bd := f(&p.group)
	p.pl = append(p.pl, bson.D{{tmorm.GroupOp, bd}})
	return p
}

func (p *Pipeline) Sort(keys ...string) *Pipeline {
	var res bson.D = make(bson.D, 0, len(keys))
	for _, k := range keys {
		res = append(res, bson.E{k, 1})
	}
	p.pl = append(p.pl, bson.D{{tmorm.SortOp, res}})
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

	p.pl = append(p.pl, bson.D{{tmorm.ProjectOp, res}})
	return p
}

func (p *Pipeline) AppendRaw(v bson.D) *Pipeline {
	p.pl = append(p.pl, v)
	return p
}
