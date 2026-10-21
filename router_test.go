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
	// Чистые функции без всякого Wrap снаружи.
	pr.GET("/users", func(*gin.Context) (any, error) { return []string{"a"}, nil })
	pr.POST("/users", func(*gin.Context) (any, error) {
		return gin.H{"id": "1"}, nil
	}, http.StatusCreated)
	api := pr.Group("/api")
	api.GET("/none", func(*gin.Context) (any, error) { return nil, nil })

	cases := []struct {
		method, path string
		want         int
		body         string
	}{
		{http.MethodGet, "/users", 200, `{"data":["a"]}`},
		{http.MethodPost, "/users", 201, `{"data":{"id":"1"}}`},
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
