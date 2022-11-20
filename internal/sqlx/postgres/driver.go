package postgres

import (
	"database/sql/driver"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/zzztttkkk/0.0/internal/sqlx"
)

type Driver struct{}

func (_ *Driver) Open(dsn string) (driver.Connector, error) {
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
