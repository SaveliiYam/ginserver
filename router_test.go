package ginserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
)

func TestRouterAcceptsPureHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	pr := ginserver.NewRouter(e)
	pr.GET("/users", func(*gin.Context) ginserver.Result { return ginserver.Data([]string{"a"}) })
	api := pr.Group("/api")
	api.GET("/none", func(*gin.Context) ginserver.Result { return ginserver.Empty() })

	cases := []struct {
		method, path string
		want         int
		body         string
	}{
		{http.MethodGet, "/users", 200, `{"data":["a"]}`},
		{http.MethodGet, "/api/none", 204, ""},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.want || w.Body.String() != tc.body {
			t.Fatalf("%s %s = %d %q, want %d %q", tc.method, tc.path, w.Code, w.Body.String(), tc.want, tc.body)
		}
	}
}
