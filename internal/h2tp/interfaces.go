package h2tp

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

type RequestCtx struct {
	Request    *http.Request
	PathParams httprouter.Params
	http.ResponseWriter
}

func (rctx *RequestCtx) Deadline() (deadline time.Time, ok bool) {
	return rctx.Request.Context().Deadline()
}

func (rctx *RequestCtx) Done() <-chan struct{} {
	return rctx.Request.Context().Done()
}

func (rctx *RequestCtx) Err() error {
	return rctx.Request.Context().Err()
}

func (rctx *RequestCtx) Value(key any) any {
	return rctx.Request.Context().Value(key)
}

var _ context.Context = (*RequestCtx)(nil)

type Handler interface {
	Handle(rctx *RequestCtx)
}

type HandlerFunc func(rctx *RequestCtx)

func (f HandlerFunc) Handle(rctx *RequestCtx) {
	f(rctx)
}

type Middleware interface {
	Handle(rctx *RequestCtx, next func())
}
