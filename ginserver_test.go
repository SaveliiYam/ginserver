package ginserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/httpx"
)

func TestHealthAndErrorFormat(t *testing.T) {
	srv := ginserver.New(ginserver.Config{Mode: gin.TestMode}, func(r gin.IRouter) {
		ginserver.Health(r)
		r.GET("/fail", func(c *gin.Context) { ginserver.Error(c, httpx.NotFound("gone")) })
	})
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("healthz = %d", w.Code)
	}
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/fail", nil))
	if w.Code != http.StatusNotFound || w.Body.String() == "" {
		t.Fatalf("fail = %d %s", w.Code, w.Body.String())
	}
}
