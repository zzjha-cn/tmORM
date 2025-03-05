package tmorm

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type (
	MDB struct {
		cli *mongo.Client
	}

	MSession struct {
		Ctx        context.Context
		mdb        *MDB
		DBName     string
		collection string
	}
)

func NewMDB(cli *mongo.Client) *MDB {
	return &MDB{cli: cli}
}

func (m *MDB) Sess(ctx context.Context, db, collection string) MSession {
	return MSession{
		Ctx:        ctx,
		mdb:        m,
		DBName:     db,
		collection: collection,
	}
}

// ======================================================

func (s MSession) Conn() *mongo.Collection {
	return s.mdb.cli.Database(s.DBName).Collection(s.collection)
}

func (s MSession) Before() {
}
