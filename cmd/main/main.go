package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"github.com/zzztttkkk/0.0/config"
	"github.com/zzztttkkk/0.0/internal"
	"github.com/zzztttkkk/0.0/internal/h2tp"
	"github.com/zzztttkkk/0.0/internal/utils"
)

const (
	DefaultConfigPath = "./.0.0.config.toml"
	LocalConfigPath   = "./.0.0.config.local.toml"
)

//go:generate go run ../autoload
func main() {
	var conf config.Config

	if utils.FsExists(DefaultConfigPath) {
		var temp config.Config
		if _, err := toml.DecodeFile(DefaultConfigPath, &temp); err != nil {
			panic(err)
		}
		if err := mergo.Merge(&conf, &temp); err != nil {
			panic(err)
		}
	}

	if utils.FsExists(LocalConfigPath) {
		var temp config.Config
		if _, err := toml.DecodeFile(LocalConfigPath, &temp); err != nil {
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

	internal.InvokeAll(5)

	fmt.Printf("Pid: %d\r\n", os.Getpid())

	for range ctx.Done() {
		break
	}
}
