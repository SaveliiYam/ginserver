package ginserver_test

import (
	"errors"
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

func engineWith(path string, h ginserver.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.GET(path, ginserver.Wrap(h))
	return e
}

func TestResultUtilities(t *testing.T) {
	// 200 объект: только data
	e := engineWith("/obj", func(*gin.Context) ginserver.Result { return ginserver.Data(gin.H{"id": "1"}) })
	if w := serve(e, "/obj"); w.Code != 200 || w.Body.String() != `{"data":{"id":"1"}}` {
		t.Fatalf("obj: %d %s", w.Code, w.Body)
	}
	// 200 пустой массив сохраняется
	e = engineWith("/arr", func(*gin.Context) ginserver.Result { return ginserver.Data([]string{}) })
	if w := serve(e, "/arr"); w.Body.String() != `{"data":[]}` {
		t.Fatalf("arr: %s", w.Body)
	}
	// 204 без тела
	e = engineWith("/none", func(*gin.Context) ginserver.Result { return ginserver.Empty() })
	if w := serve(e, "/none"); w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("none: %d %q", w.Code, w.Body)
	}
	// 401
	e = engineWith("/unauth", func(*gin.Context) ginserver.Result {
		return ginserver.Unauthorized(errors.New("bad token"))
	})
	if w := serve(e, "/unauth"); w.Code != 401 || w.Body.String() != `{"errors":[{"code":"unauthorized","message":"bad token"}]}` {
		t.Fatalf("unauth: %d %s", w.Code, w.Body)
	}
	// 422 берёт сообщение из ошибки
	e = engineWith("/unp", func(*gin.Context) ginserver.Result {
		return ginserver.Unprocessable(errors.New("name is required"))
	})
	if w := serve(e, "/unp"); w.Code != 422 || w.Body.String() != `{"errors":[{"code":"unprocessable_entity","message":"name is required"}]}` {
		t.Fatalf("unp: %d %s", w.Code, w.Body)
	}
	// 500 не светит внутреннюю ошибку
	e = engineWith("/int", func(*gin.Context) ginserver.Result {
		return ginserver.Internal(errors.New("db connection refused at 10.0.0.5"))
	})
	if w := serve(e, "/int"); w.Code != 500 || w.Body.String() != `{"errors":[{"code":"internal","message":"internal error"}]}` {
		t.Fatalf("int: %d %s", w.Code, w.Body)
	}
	// httpx-ошибка внутри Unprocessable отдаёт свой код
	e = engineWith("/unp2", func(*gin.Context) ginserver.Result {
		return ginserver.Unprocessable(httpx.NotFound("user not found"))
	})
	if w := serve(e, "/unp2"); w.Code != 422 || w.Body.String() != `{"errors":[{"code":"not_found","message":"user not found"}]}` {
		t.Fatalf("unp2: %d %s", w.Code, w.Body)
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
