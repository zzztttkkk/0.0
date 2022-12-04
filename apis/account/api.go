package account

import (
	"github.com/zzztttkkk/0.0/config"
	"github.com/zzztttkkk/0.0/internal"
	"github.com/zzztttkkk/0.0/internal/h2tp"
)

type AutoExport struct{}

func init() {
	internal.LazyInvoke(func(cfg *config.Config, router *h2tp.Router) {
		router.Register("post", "", h2tp.HandlerFunc(func(rctx *h2tp.RequestCtx) {

		}))
	})
}
