package main

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"github.com/zzztttkkk/0.0/config"
	"github.com/zzztttkkk/0.0/internal"
	"github.com/zzztttkkk/0.0/internal/h2tp"
	"github.com/zzztttkkk/0.0/internal/utils"
	"os"
	"os/signal"
)

func main() {
	var conf config.Config

	if utils.FsExists("./.config.toml") {
		var temp config.Config
		if _, err := toml.DecodeFile("./.config.toml", &temp); err != nil {
			panic(err)
		}
		if err := mergo.Merge(&conf, &temp); err != nil {
			panic(err)
		}
	}

	if utils.FsExists("./.config.local.toml") {
		var temp config.Config
		if _, err := toml.DecodeFile("./.config.local.toml", &temp); err != nil {
			panic(err)
		}
		if err := mergo.Merge(&conf, &temp, mergo.WithOverride); err != nil {
			panic(err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf.Init(ctx)

	internal.Provide(func() *config.Config { return &conf })

	internal.Provide(func() *h2tp.Router { return h2tp.NewRouter() })

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
