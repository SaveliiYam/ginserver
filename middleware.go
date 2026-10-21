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

// RequestLogger — обязательный логгер каждого запроса: одна запись на запрос
// после его завершения. Ставится сервером всегда (см. New), отключить нельзя;
// настраивается только сам логгер (WithLogger).
//
// Поля: method, path (сырой), route (шаблон gin, если совпал), status,
// duration_ms, bytes (размер ответа), client_ip, request_id, errors (если есть).
// Уровень: 5xx -> error, 4xx -> warn, иначе info.
func RequestLogger(log logger.Logger) gin.HandlerFunc {
	if log == nil {
		log = logger.Nop{}
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		size := c.Writer.Size()
		if size < 0 { // gin возвращает -1, если тело не писали
			size = 0
		}
		args := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", c.FullPath(),
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", size,
			"client_ip", c.ClientIP(),
			"request_id", httpx.RequestID(c.Request.Context()),
		}
		if len(c.Errors) > 0 {
			args = append(args, "errors", c.Errors.String())
		}
		ctx := c.Request.Context()
		switch {
		case status >= 500:
			log.Error(ctx, "http request", args...)
		case status >= 400:
			log.Warn(ctx, "http request", args...)
		default:
			log.Info(ctx, "http request", args...)
		}
	}
}

// AccessLog — устаревший алиас RequestLogger (совместимость с v0.1.x–v0.4.x).
//
// Deprecated: используйте RequestLogger.
func AccessLog(log logger.Logger) gin.HandlerFunc { return RequestLogger(log) }

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
