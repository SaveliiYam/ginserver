package ginserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Router — gin-роутер, принимающий чистые HandlerFunc, возвращающие Result.
// Оборачивание (Wrap) делается внутри: в прикладном коде обёрток нет,
// статус обработчик задаёт сам утилитами Data/Empty/Unauthorized/...:
//
//	pr := ginserver.NewRouter(engine)
//	pr.GET("/users", handlers.List)
//	api := pr.Group("/api/v1")
type Router struct {
	r gin.IRouter
}

func NewRouter(r gin.IRouter) *Router { return &Router{r: r} }

// Group возвращает Router для группы маршрутов.
func (rt *Router) Group(prefix string, mw ...gin.HandlerFunc) *Router {
	switch r := rt.r.(type) {
	case *gin.RouterGroup:
		return &Router{r: r.Group(prefix, mw...)}
	case *gin.Engine:
		return &Router{r: r.Group(prefix, mw...)}
	default:
		panic("ginserver: Router.Group: unsupported gin.IRouter implementation")
	}
}

// Handle регистрирует чистый обработчик на method+path.
func (rt *Router) Handle(method, path string, h HandlerFunc) {
	rt.r.Handle(method, path, Wrap(h))
}

func (rt *Router) GET(path string, h HandlerFunc)    { rt.Handle(http.MethodGet, path, h) }
func (rt *Router) POST(path string, h HandlerFunc)   { rt.Handle(http.MethodPost, path, h) }
func (rt *Router) PUT(path string, h HandlerFunc)    { rt.Handle(http.MethodPut, path, h) }
func (rt *Router) PATCH(path string, h HandlerFunc)  { rt.Handle(http.MethodPatch, path, h) }
func (rt *Router) DELETE(path string, h HandlerFunc) { rt.Handle(http.MethodDelete, path, h) }
