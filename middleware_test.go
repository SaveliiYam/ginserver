package ginserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/logger"
)

type entry struct {
	level string
	msg   string
	args  []any
}

type fakeLogger struct {
	mu      sync.Mutex
	entries []entry
}

func (f *fakeLogger) add(level, msg string, args []any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, entry{level, msg, args})
}
func (f *fakeLogger) Debug(_ context.Context, msg string, args ...any) { f.add("debug", msg, args) }
func (f *fakeLogger) Info(_ context.Context, msg string, args ...any)  { f.add("info", msg, args) }
func (f *fakeLogger) Warn(_ context.Context, msg string, args ...any)  { f.add("warn", msg, args) }
func (f *fakeLogger) Error(_ context.Context, msg string, args ...any) { f.add("error", msg, args) }
func (f *fakeLogger) With(...any) logger.Logger                        { return f }

func field(e entry, key string) any {
	for i := 0; i+1 < len(e.args); i += 2 {
		if e.args[i] == key {
			return e.args[i+1]
		}
	}
	return nil
}

func TestRequestLoggerIsMandatory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fl := &fakeLogger{}
	srv := ginserver.New(ginserver.Config{Mode: gin.TestMode}, func(r gin.IRouter) {
		r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/bad", func(c *gin.Context) { c.Status(http.StatusUnprocessableEntity) })
		r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
	}, ginserver.WithLogger(fl))

	for _, path := range []string{"/ok", "/bad", "/boom", "/ping", "/missing"} {
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	}

	fl.mu.Lock()
	defer fl.mu.Unlock()
	if len(fl.entries) != 5 {
		t.Fatalf("entries = %d, want 5 (every request logged): %+v", len(fl.entries), fl.entries)
	}
	wantLevel := []string{"info", "warn", "error", "info", "warn"}
	for i, e := range fl.entries {
		if e.msg != "http request" || e.level != wantLevel[i] {
			t.Fatalf("entry %d = %s/%s, want http request/%s", i, e.level, e.msg, wantLevel[i])
		}
		if field(e, "request_id") == nil || field(e, "status") == nil || field(e, "method") != "GET" {
			t.Fatalf("entry %d missing fields: %+v", i, e.args)
		}
	}
	if got := field(fl.entries[0], "route"); got != "/ok" {
		t.Fatalf("route = %v, want /ok", got)
	}
}
