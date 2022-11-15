package h2tp

import (
	"net/http"
	"testing"
)

func TestRun(t *testing.T) {
	router := NewRouter()
	router.Register(http.MethodGet, "/spk", HandlerFunc(func(rctx *RequestCtx) {
		rctx.WriteHeader(200)
		_, _ = rctx.Write([]byte("Hello World"))
	}))

	if e := Run("127.0.0.1:8524", map[string]*Router{"*": router}); e != nil {
		panic(e)
	}
}
