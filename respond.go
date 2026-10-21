package ginserver

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/httpx"
)

// OK пишет 200 JSON.
func OK(c *gin.Context, v any) { c.JSON(http.StatusOK, v) }

// Created пишет 201 JSON.
func Created(c *gin.Context, v any) { c.JSON(http.StatusCreated, v) }

// NoContent пишет 204.
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// Error пишет ошибку в едином формате httpx: {"error":{"code","message"}}.
// Внутренняя причина наружу не попадает — залогируйте её до вызова.
func Error(c *gin.Context, err error) {
	e := httpx.AsError(err)
	c.JSON(e.Status, gin.H{"error": gin.H{"code": e.Code, "message": e.Message}})
}

// BindJSON биндит тело; при ошибке сразу пишет 400 и возвращает false.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		Error(c, httpx.BadRequest("invalid json body"))
		return false
	}
	return true
}

// Page — параметры пагинации (?page=1&page_size=20, максимум 100).
type Page struct {
	Number int `form:"page"`
	Size   int `form:"page_size"`
}

func (p Page) Limit() int  { return p.Size }
func (p Page) Offset() int { return (p.Number - 1) * p.Size }

// ParsePage читает пагинацию с дефолтами 1/20.
func ParsePage(c *gin.Context) Page {
	n, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	s, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if n < 1 {
		n = 1
	}
	if s < 1 {
		s = 20
	}
	if s > 100 {
		s = 100
	}
	return Page{Number: n, Size: s}
}

// Health регистрирует GET /healthz (и /readyz, если заданы проверки готовности).
func Health(r gin.IRouter, ready ...func(context.Context) error) {
	r.GET("/healthz", func(c *gin.Context) { OK(c, gin.H{"status": "ok"}) })
	if len(ready) > 0 {
		r.GET("/readyz", func(c *gin.Context) {
			for _, check := range ready {
				if err := check(c.Request.Context()); err != nil {
					Error(c, httpx.NewError(http.StatusServiceUnavailable, "not_ready", "not ready"))
					return
				}
			}
			OK(c, gin.H{"status": "ok"})
		})
	}
}
