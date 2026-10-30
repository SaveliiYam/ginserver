package ginserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoadAccountsJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "basic_auth.json")
	if err := os.WriteFile(path, []byte(`{"a":"1","b":"2"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	acc, err := LoadAccountsJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if acc["a"] != "1" || acc["b"] != "2" {
		t.Fatalf("%v", acc)
	}
}

func TestBasicAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BasicAuth(Accounts{"u": "p"}, WithRealm("jwt")))
	r.GET("/srv/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	t.Run("ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/srv/x", nil)
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("u:p")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/srv/x", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", w.Code)
		}
		var body EnvelopeJSON
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Errors) == 0 {
			t.Fatalf("body=%s", w.Body.String())
		}
	})

	t.Run("skip healthz", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d", w.Code)
		}
	})
}

func TestBasicAuthEmptyAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BasicAuth(nil))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}
