package ginserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Accounts — словарь login → password для HTTP Basic Auth.
type Accounts map[string]string

// BasicAuthOption настраивает BasicAuth.
type BasicAuthOption func(*basicAuthOptions)

type basicAuthOptions struct {
	realm     string
	skipPaths map[string]struct{}
}

// WithRealm задаёт WWW-Authenticate realm (по умолчанию "restricted").
func WithRealm(realm string) BasicAuthOption {
	return func(o *basicAuthOptions) {
		if realm != "" {
			o.realm = realm
		}
	}
}

// WithSkipPaths пути без Basic (по умолчанию /healthz и /ping).
func WithSkipPaths(paths ...string) BasicAuthOption {
	return func(o *basicAuthOptions) {
		o.skipPaths = make(map[string]struct{}, len(paths))
		for _, p := range paths {
			o.skipPaths[p] = struct{}{}
		}
	}
}

// LoadAccountsJSON читает файл вида {"login1":"password1","login2":"password2"}.
func LoadAccountsJSON(path string) (Accounts, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("basicauth: read %s: %w", path, err)
	}
	var acc Accounts
	if err := json.Unmarshal(data, &acc); err != nil {
		return nil, fmt.Errorf("basicauth: parse %s: %w", path, err)
	}
	if acc == nil {
		acc = Accounts{}
	}
	return acc, nil
}

// BasicAuth — middleware: пропускает запрос, если Authorization Basic совпал
// с одной из пар в accounts. Пустой accounts — no-op (auth выключен).
//
// При ошибке пишет конверт ginserver Unauthorized и Abort.
func BasicAuth(accounts Accounts, opts ...BasicAuthOption) gin.HandlerFunc {
	o := basicAuthOptions{
		realm: "restricted",
		skipPaths: map[string]struct{}{
			PingPath:   {},
			"/healthz": {},
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	if len(accounts) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	realmHeader := fmt.Sprintf(`Basic realm=%q`, o.realm)

	return func(c *gin.Context) {
		if _, skip := o.skipPaths[c.Request.URL.Path]; skip {
			c.Next()
			return
		}

		user, pass, ok := parseBasicAuthHeader(c.GetHeader("Authorization"))
		if !ok {
			unauthorized(c, realmHeader)
			return
		}
		want, exists := accounts[user]
		if !exists || want != pass {
			unauthorized(c, realmHeader)
			return
		}
		c.Set("basic_auth_user", user)
		c.Next()
	}
}

func unauthorized(c *gin.Context, realmHeader string) {
	c.Header("WWW-Authenticate", realmHeader)
	res := Unauthorized("basic-auth-unauthorized", nil)
	c.AbortWithStatusJSON(res.Status, res.Body)
}

func parseBasicAuthHeader(header string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[len(prefix):]))
	if err != nil {
		return "", "", false
	}
	username, password, ok = strings.Cut(string(raw), ":")
	return username, password, ok
}
