package config

import (
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"github.com/zzztttkkk/0.0/internal/sqlx/postgres"
)

func (cfg *Config) DBGroup() *sqlx.Group { return cfg.internal.group }

func (cfg *Config) initDb() {
	dbcfg := cfg.Database
	gopts := &sqlx.GroupOptions{
		ReadonlySourceNames: dbcfg.Slavers,
		MaxIdleConns:        dbcfg.MaxIdleConns,
		MaxOpenConns:        dbcfg.MaxOpenConns,
		ConnMaxIdleTime:     dbcfg.ConnMaxIdleTime,
		ConnMaxLifetime:     dbcfg.ConnMaxLifetime,
	}
	cfg.internal.group = sqlx.NewGroup(&postgres.Driver{}, dbcfg.Master, gopts)
}
