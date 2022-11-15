package h2tp

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type Router struct {
	frozen bool

	internal   *httprouter.Router
	middleware []Middleware
}

func NewRouter() *Router {
	obj := &Router{}
	obj.internal = httprouter.New()
	return obj
}

func (r *Router) Use(middleware Middleware) {
	r.mustBeModifiable()

	r.middleware = append(r.middleware, middleware)
}

func (r *Router) makeHandler(handler Handler) Handler {
	return HandlerFunc(func(rctx *RequestCtx) {
		var idx = -1
		var next func()
		next = func() {
			idx++
			if idx < len(r.middleware) {
				r.middleware[idx].Handle(rctx, next)
			} else {
				handler.Handle(rctx)
			}
		}
		next()
	})
}

func (r *Router) mustBeModifiable() {
	if r.frozen {
		panic("router is already frozen")
	}
}

func (r *Router) Register(method string, pattern string, handler Handler) {
	r.mustBeModifiable()

	r.internal.Handle(method, pattern, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		rctx := RequestCtx{
			Request:        request,
			ResponseWriter: writer,
			PathParams:     params,
		}
		handler = r.makeHandler(handler)
		handler.Handle(&rctx)
	})
}

func Run(addr string, routers map[string]*Router) error {
	for _, router := range routers {
		router.frozen = true
	}

	var defaultRouter *Router
	for _, key := range []string{"", "*", "default"} {
		if defaultRouter != nil {
			break
		}
		defaultRouter = routers[key]
	}
	return http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		router := routers[request.Host]
		if router == nil {
			router = defaultRouter
		}

		if router == nil {
			return
		}

		fn, params, _ := router.internal.Lookup(request.Method, request.RequestURI)
		if fn == nil {
			return
		}

		fn(writer, request, params)
	}))
}
