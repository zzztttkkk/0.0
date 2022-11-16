package h2tp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/zzztttkkk/0.0/internal/utils"
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

func (r *Router) makeMiddlewareWrapper(handler Handler) Handler {
	return HandlerFunc(func(rctx *RequestCtx) {
		var next func()
		next = func() {
			rctx.middlewareIdx++
			if rctx.middlewareIdx < len(r.middleware) {
				r.middleware[rctx.middlewareIdx].Handle(rctx, next)
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

var AllMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

func (r *Router) Register(methods string, pattern string, handler Handler) {
	r.mustBeModifiable()

	var temp []string
	if methods == "*" {
		temp = AllMethods
	} else {
		for _, part := range strings.Split(methods, ",") {
			part = strings.ToUpper(strings.TrimSpace(part))
			if utils.SliceFind(AllMethods, part) < 0 {
				panic(fmt.Errorf("unknown method, %s", part))
			}
			temp = append(temp, part)
		}
	}

	for _, method := range temp {
		r.internal.Handle(method, pattern, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
			rctx := RequestCtx{
				Request:        request,
				ResponseWriter: writer,
				PathParams:     params,
				middlewareIdx:  -1,
			}
			handler = r.makeMiddlewareWrapper(handler)
			handler.Handle(&rctx)
		})
	}
}

func Run(addr string, routers map[string]*Router) error {
	if len(routers) < 1 {
		routers = map[string]*Router{}
		router := NewRouter()
		routers["*"] = router
	}

	for _, router := range routers {
		router.frozen = true
	}

	defaultRouter := routers["*"]
	delete(routers, "*")

	var peekRouter func(r *http.Request) *Router
	if len(routers) == 0 {
		peekRouter = func(_ *http.Request) *Router {
			return defaultRouter
		}
	} else {
		if len(routers) == 1 {
			host := utils.MapKeys(routers)[0]
			router := utils.MapValues(routers)[0]
			peekRouter = func(req *http.Request) *Router {
				if req.Host != host {
					return nil
				}
				return router
			}
		} else {
			if len(routers) < 6 {
				_hosts := utils.MapKeys(routers)
				_routers := utils.MapValues(routers)
				peekRouter = func(r *http.Request) *Router {
					for i, host := range _hosts {
						if r.Host == host {
							return _routers[i]
						}
					}
					return nil
				}
			} else {
				peekRouter = func(r *http.Request) *Router {
					return routers[r.Host]
				}
			}
		}
	}

	return http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		router := peekRouter(request)
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
