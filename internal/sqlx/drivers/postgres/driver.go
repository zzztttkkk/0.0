package postgres

import (
	"database/sql/driver"
	"fmt"
	"github.com/lib/pq"
	"github.com/zzztttkkk/0.0/internal/sqlx/internal"
)

type Driver struct{}

func (my *Driver) Open(dsn string) (driver.Connector, error) { return pq.NewConnector(dsn) }

func (my *Driver) Placeholder(idx int, _ string) string { return fmt.Sprintf("$%d", idx) }

var (
	_ internal.Driver = (*Driver)(nil)
)
