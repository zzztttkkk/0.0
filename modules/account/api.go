package account

import (
	"github.com/zzztttkkk/0.0/config"
	"github.com/zzztttkkk/0.0/internal"
	"github.com/zzztttkkk/0.0/internal/h2tp"
	"net/http"
)

var (
	gcfg *config.Config
)

func httpRegister(rctx *h2tp.RequestCtx) {

}

func init() {
	internal.Invoke(func(cfg *config.Config, router *h2tp.Router) {
		gcfg = cfg

		router.Register(http.MethodPost, "/api/account/register", h2tp.HandlerFunc(httpRegister))
	})
}
