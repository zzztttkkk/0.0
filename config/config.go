package config

import (
	"context"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"github.com/zzztttkkk/0.0/internal/sqlx/postgres"
)

type Config struct {
	internal struct {
		ctx    context.Context
		group  *sqlx.Group
		master *postgres.DB
	} `toml:"-"`

	Database struct {
		Master          string   `toml:"master"`
		Slavers         []string `toml:"slavers"`
		MaxIdleConns    int      `toml:"max_idle_conns"`
		MaxOpenConns    int      `toml:"max_open_conns"`
		ConnMaxLifetime int      `toml:"conn_max_lifetime"`
		ConnMaxIdleTime int      `toml:"conn_max_idle_time"`
	} `toml:"database"`

	Redis struct {
		Data  string `toml:"data"`
		Cache string `toml:"cache"`
	} `toml:"redis"`
}

func (cfg *Config) Init(ctx context.Context) {
	cfg.internal.ctx = ctx
	cfg.initDb()
}

func (cfg *Config) Context() context.Context { return cfg.internal.ctx }
