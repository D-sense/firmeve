package http

import (
	"github.com/firmeve/firmeve/kernel"
	"github.com/firmeve/firmeve/kernel/contract"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

type Router struct {
	Firmeve   contract.Application
	router    *httprouter.Router
	routes    map[string]*Route
	routeKeys []string
}

func New(firmeve contract.Application) *Router {
	return &Router{
		Firmeve:   firmeve,
		router:    httprouter.New(),
		routes:    make(map[string]*Route, 0),
		routeKeys: make([]string, 0),
	}
}

func (r *Router) GET(path string, handler contract.ContextHandler) *Route {
	return r.createRoute(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler contract.ContextHandler) *Route {
	return r.createRoute(http.MethodPost, path, handler)
}

func (r *Router) PUT(path string, handler contract.ContextHandler) *Route {
	return r.createRoute(http.MethodPut, path, handler)
}

func (r *Router) PATCH(path string, handler contract.ContextHandler) *Route {
	return r.createRoute(http.MethodPatch, path, handler)
}

func (r *Router) DELETE(path string, handler contract.ContextHandler) *Route {
	return r.createRoute(http.MethodDelete, path, handler)
}

func (r *Router) OPTIONS(path string, handler contract.ContextHandler) *Route {
	return r.createRoute(http.MethodOptions, path, handler)
}

// serve static files
func (r *Router) Static(path string, root string) *Router {
	r.router.ServeFiles(strings.Join([]string{path, `/*filepath`}, ``), http.Dir(root))
	return r
}

func (r *Router) NotFound(handler contract.ContextHandler) *Router {
	r.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		kernel.NewContext(r.Firmeve, NewHttp(req, w), handler).Next()
		//newContext(r.Firmeve, w, req, handler).Next()
	})
	return r
}

func (r *Router) Handler(method, path string, handler http.HandlerFunc) {
	r.createRoute(method, path, func(c contract.Context) {
		protocol := c.Protocol().(contract.HttpProtocol)
		handler(protocol.ResponseWriter(), protocol.Request())
	})
}

func (r *Router) HttpRouter() *httprouter.Router {
	return r.router
}

//
func (r *Router) Group(prefix string) *Group {
	return newGroup(r).Prefix(prefix)
}

func (r *Router) createRoute(method string, path string, handler contract.ContextHandler) *Route {
	key := r.routeKey(method, path)
	r.routes[key] = newRoute(path, handler)

	//Only http router
	//r.router.Handler(method, path, r)
	r.router.Handle(method, path, func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctxParams := make(map[string]string, 0)
		for _, param := range params {
			ctxParams[param.Key] = param.Value
		}

		ctx := kernel.NewContext(r.Firmeve, NewHttp(req, w), r.routes[key].Handlers()...)
		//ctx := newContext(r.Firmeve, w, req, r.routes[key].Handlers()...).
		//	SetParams(ctxParams).
		//	SetRoute(r.routes[key])

		r.Firmeve.Get(`event`).(contract.Event).Dispatch(`router.match`, map[string]interface{}{
			`context`: ctx,
			`route`:   r.routes[key],
		})

		ctx.Next()
	})

	return r.routes[key]
}

func (r *Router) routeKey(method, path string) string {
	return strings.Join([]string{method, path}, `.`)
}

// 只是用到httpRouter的存储路由以及查找方法，因为暂时不会前缀树算法
// 其它router,middleware这些都是我自己实现，惟一对接的就是无缝的写入一套httprouter规则的路由（后期替换为自己的路由）
// 通过ServerHttp去查找匹配路由
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	r.router.ServeHTTP(w, req)

	// r.routes[r.routeKey(req.Method, req.URL.Path)]也可以使用
	// r.router.Lookup(req.Method, req.URL.Path+"/") 来实现
	//if _, ok := r.routes[r.routeKey(req.Method, req.URL.Path)]; !ok && r.notFound != nil {
	//	newContext(w, req, r.notFound).Next()
	//} else {
	//	r.router.ServeHTTP(w, req)
	//}
}
