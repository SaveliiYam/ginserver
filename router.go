package ginserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Router — gin-роутер, принимающий чистые HandlerFunc: Wrap делается внутри,
// в прикладном коде обёртку писать не нужно.
//
//	pr := ginserver.NewRouter(engine)
//	pr.GET("/users", handlers.List)
//	pr.POST("/users", handlers.Create, http.StatusCreated) // необязательный статус
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
// status — необязательный статус успеха (по умолчанию 200).
func (rt *Router) Handle(method, path string, h HandlerFunc, status ...int) {
	rt.r.Handle(method, path, WrapStatus(pickStatus(status), h))
}

func (rt *Router) GET(path string, h HandlerFunc, status ...int) {
	rt.Handle(http.MethodGet, path, h, status...)
}
func (rt *Router) POST(path string, h HandlerFunc, status ...int) {
	rt.Handle(http.MethodPost, path, h, status...)
}
func (rt *Router) PUT(path string, h HandlerFunc, status ...int) {
	rt.Handle(http.MethodPut, path, h, status...)
}
func (rt *Router) PATCH(path string, h HandlerFunc, status ...int) {
	rt.Handle(http.MethodPatch, path, h, status...)
}
func (rt *Router) DELETE(path string, h HandlerFunc, status ...int) {
	rt.Handle(http.MethodDelete, path, h, status...)
}

func pickStatus(status []int) int {
	if len(status) > 0 && status[0] > 0 {
		return status[0]
	}
	return http.StatusOK
}
