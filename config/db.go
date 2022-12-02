package config

import (
	"context"
	"github.com/ulule/deepcopier"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"github.com/zzztttkkk/0.0/internal/sqlx/postgres"
	"time"
)

func (cfg *Config) DBGroup() *sqlx.Group { return cfg.internal.group }

func (cfg *Config) DBMaster() *postgres.DB { return cfg.internal.master }

func (cfg *Config) initDb() {
	dbcfg := cfg.Database
	gopts := &sqlx.GroupOptions{ReadonlySourceNames: dbcfg.Slavers}
	_ = deepcopier.Copy(dbcfg).To(gopts)
	cfg.internal.group = sqlx.NewGroup(&postgres.Driver{}, dbcfg.Master, gopts)
	cfg.internal.master = &postgres.DB{DB: cfg.internal.group.DB()}

	ctx, cancel := context.WithTimeout(cfg.Context(), time.Second*5)
	defer cancel()

	cfg.internal.master.EnableCrypto(ctx).EnableHStore(ctx).EnableUUID(ctx)
}
