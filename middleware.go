package ginserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/httpx"
	"github.com/SaveliiYam/logger"
)

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}

// RequestID ставит X-Request-Id (входящий или новый UUID) и кладёт его в контекст запроса.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}
		c.Header("X-Request-Id", id)
		c.Request = c.Request.WithContext(httpx.WithRequestID(c.Request.Context(), id))
		c.Set("request_id", id)
		c.Next()
	}
}

// Recovery ловит паники -> 500 в едином формате, пишет стект в лог.
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error(c.Request.Context(), "panic recovered", "panic", r, "path", c.Request.URL.Path)
				Error(c, httpx.Internal(errors.New("panic")))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// AccessLog — одна строка на запрос.
func AccessLog(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info(c.Request.Context(), "http request",
			"method", c.Request.Method, "path", c.Request.URL.Path,
			"status", c.Writer.Status(), "dur", time.Since(start).String(),
			"request_id", httpx.RequestID(c.Request.Context()))
	}
}

// CORS — простой CORS. Для прод-сценариев со сложной политикой ставьте свой middleware через WithMiddleware.
func CORS(allowOrigins ...string) gin.HandlerFunc {
	origin := "*"
	if len(allowOrigins) > 0 {
		origin = allowOrigins[0]
	}
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-Id")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Timeout ограничивает время хендлера через контекст запроса.
// Внимание: хендлер обязан уважать c.Request.Context().
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
