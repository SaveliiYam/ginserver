package ginserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/httpx"
)

func serve(e *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func engineWith(path string, h gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.GET(path, h)
	return e
}

func TestWrapOmitsEmptyFields(t *testing.T) {
	// объект: только data, без errors и без null
	e := engineWith("/obj", ginserver.Wrap(func(*gin.Context) (any, error) {
		return gin.H{"id": "1"}, nil
	}))
	if body := serve(e, "/obj").Body.String(); body != "{\"data\":{\"id\":\"1\"}}" {
		t.Fatalf("object: %s", body)
	}
	// пустой массив: data: [] сохраняется
	e = engineWith("/arr", ginserver.Wrap(func(*gin.Context) (any, error) {
		return []string{}, nil
	}))
	if body := serve(e, "/arr").Body.String(); body != "{\"data\":[]}" {
		t.Fatalf("empty array: %s", body)
	}
	// ошибка: только errors, без data
	e = engineWith("/err", ginserver.Wrap(func(*gin.Context) (any, error) {
		return nil, httpx.NotFound("user not found")
	}))
	w := serve(e, "/err")
	var env map[string]json.RawMessage
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if w.Code != 404 {
		t.Fatalf("err code = %d", w.Code)
	}
	if _, hasData := env["data"]; hasData {
		t.Fatalf("error body must omit data: %s", w.Body)
	}
	if _, hasErrors := env["errors"]; !hasErrors {
		t.Fatalf("error body must have errors: %s", w.Body)
	}
	// нет данных: 204 без тела
	e = engineWith("/none", ginserver.Wrap(func(*gin.Context) (any, error) { return nil, nil }))
	w = serve(e, "/none")
	if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("no content: %d %q", w.Code, w.Body.String())
	}
}

func TestDefaultPingAndOptOut(t *testing.T) {
	srv := ginserver.New(ginserver.Config{Mode: gin.TestMode}, nil)
	w := serve(srv.Engine(), "/ping")
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("ping: %d %q", w.Code, w.Body.String())
	}
	srv = ginserver.New(ginserver.Config{Mode: gin.TestMode}, nil, ginserver.WithoutPing())
	if w := serve(srv.Engine(), "/ping"); w.Code != 404 {
		t.Fatalf("ping disabled: %d", w.Code)
	}
}
