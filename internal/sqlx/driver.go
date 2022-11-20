package sqlx

import (
	"database/sql/driver"
	"github.com/zzztttkkk/0.0/internal/utils"
)

type Driver interface {
	Open(dsn string) (driver.Connector, error)
	Placeholder(idx int, name string) string
	DDL(info *utils.FieldInfo) *FieldDefinition
}
