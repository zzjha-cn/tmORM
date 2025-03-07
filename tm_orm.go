package tmorm

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type (
	MDB struct {
		cli        *mongo.Client
		middleware *MiddleChain
	}
)

func NewMDB(cli *mongo.Client) *MDB {
	return &MDB{
		cli:        cli,
		middleware: NewMiddleChainAdapt(),
	}
}

func (m *MDB) Sess(ctx context.Context, db, collection string, msList ...MHandlerBuilder) MSession {
	return MSession{
		Ctx:        ctx,
		mdb:        m,
		DBName:     db,
		Collection: collection,
		ms:         msList,
	}
}

func (m *MDB) SetMiddleware(mc *MiddleChain) *MDB {
	m.middleware = mc
	return m
}

func (m *MDB) GetMiddleware() *MiddleChain {
	return m.middleware
}
