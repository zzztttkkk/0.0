package postgres

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/sqlx"
)

type DB struct {
	*sqlx.DB
}

func (db *DB) EnsureExtensions(ctx context.Context, exts ...string) *DB {
	if len(exts) == 0 {
		return db
	}
	for _, s := range exts {
		_, e := db.Execute(ctx, fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s";`, s), nil)
		if e != nil {
			panic(e)
		}
	}
	return db
}

func (db *DB) EnableUUID(ctx context.Context) *DB {
	return db.EnsureExtensions(ctx, "uuid-ossp")
}

func (db *DB) EnableHStore(ctx context.Context) *DB {
	return db.EnsureExtensions(ctx, "hstore")
}

func (db *DB) EnableCrypto(ctx context.Context) *DB {
	return db.EnsureExtensions(ctx, "pgcrypto")
}

func Open(dsn string, readonly bool, logger sqlx.Logger) *DB {
	return &DB{sqlx.MustOpenDB(&Driver{}, dsn, readonly, logger)}
}
