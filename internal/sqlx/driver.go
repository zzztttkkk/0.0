package sqlx

import (
	"database/sql/driver"
)

type Driver interface {
	Open(dsn string) (driver.Connector, error)
	Placeholder(idx int, name string) string
	DDL(v any) string
}
