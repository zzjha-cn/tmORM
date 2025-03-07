package tmorm

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	MSession struct {
		mdb        *MDB
		Ctx        context.Context
		DBName     string
		Collection string
		CollConn   *MDBConn
		ms         []MHandlerBuilder
	}

	MDBConn struct {
		*mongo.Collection
	}
)

func (s *MSession) Conn() *MDBConn {
	if s.CollConn != nil {
		return s.CollConn
	}
	s.CollConn = &MDBConn{}
	s.CollConn.Collection = s.mdb.cli.Database(s.DBName).Collection(s.Collection)
	return s.CollConn
}

func conn(mctx *MiddleCtx) {
	c := &MDBConn{}
	c.Collection = mctx.Session.mdb.cli.Database(mctx.Session.DBName).Collection(mctx.Session.Collection)
	mctx.Session.CollConn = c
}

func (s *MSession) BuildExecuteChain(root MHandlerFunc) MHandlerFunc {
	return s.mdb.middleware.build(root, s.ms...)
}
