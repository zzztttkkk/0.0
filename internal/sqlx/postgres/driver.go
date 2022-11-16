package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/zzztttkkk/0.0/internal/sqlx"
)

type Driver struct{}

func (my *Driver) Open(dsn string) (driver.Connector, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return stdlib.GetConnector(*cfg), nil
}

func (my *Driver) Placeholder(idx int, _ string) string { return fmt.Sprintf("$%d", idx+1) }

var (
	_ sqlx.Driver = (*Driver)(nil)
)

type DB struct {
	sqlx.DB
}

func (db *DB) EnsureExtensions(ctx context.Context, exts ...string) *DB {
	if len(exts) == 0 {
		return db
	}
	for _, s := range exts {
		_, e := db.Execute(ctx, fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s;`, s), nil)
		if e != nil {
			panic(e)
		}
	}
	return db
}
