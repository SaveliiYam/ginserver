package ginserver

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/httpx"
)

// EnvelopeJSON — глобальная структура тела любого ответа с телом:
//
//	успех: {"data": {...} | [...]}
//	ошибка: {"errors": [{"code": "...", "message": "..."}]}
//
// Data — указатель: nil опускается из JSON (omitempty), указатель на пустой
// срез даёт "data": []. Errors опускаются, когда пустые.
//
// Если данных нет и ошибки нет — тело не отправляется (204, см. Empty).
type EnvelopeJSON struct {
	Data   *any        `json:"data,omitempty"`
	Errors []ErrorBody `json:"errors,omitempty"`
}

// Envelope — алиас EnvelopeJSON для короткой записи.
type Envelope = EnvelopeJSON

// ErrorBody — одна ошибка в конверте.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Result — готовая модель ответа: статус + тело.
// Именно её возвращают обработчики (через утилиты ниже).
type Result struct {
	Status int
	Body   EnvelopeJSON
}

// HandlerFunc — обработчик, возвращающий готовый Result, ничего не пишет сам.
type HandlerFunc func(c *gin.Context) Result

// --- утилиты построения Result (стабильный набор статусов: 200, 204, 401, 422, 500) ---

// Data — 200 с данными (объект или массив).
func Data(data any) Result {
	return Result{Status: http.StatusOK, Body: EnvelopeJSON{Data: &data}}
}

// Empty — 204 без тела.
func Empty() Result {
	return Result{Status: http.StatusNoContent}
}

// Unauthorized — 401. err можно передать nil (будет общее сообщение).
func Unauthorized(err error) Result {
	return errorResult(http.StatusUnauthorized, "unauthorized", "unauthorized", err)
}

// Unprocessable — 422. err обязателен по смыслу: из него берутся код/сообщение;
// при nil подставится общее сообщение.
func Unprocessable(err error) Result {
	return errorResult(http.StatusUnprocessableEntity, "unprocessable_entity", "unprocessable entity", err)
}

// Internal — 500. err обязателен, но наружу его текст НЕ попадает:
// клиент получает общее "internal error", детализацию логируйте до возврата.
func Internal(err error) Result {
	if err == nil {
		err = errors.New("internal error")
	}
	return Result{
		Status: http.StatusInternalServerError,
		Body:   EnvelopeJSON{Errors: []ErrorBody{{Code: "internal", Message: "internal error"}}},
	}
}

func errorResult(status int, defCode, defMsg string, err error) Result {
	code, msg := defCode, defMsg
	if err != nil {
		var he *httpx.Error
		if errors.As(err, &he) {
			code, msg = he.Code, he.Message
		} else {
			msg = err.Error()
		}
	}
	return Result{
		Status: status,
		Body:   EnvelopeJSON{Errors: []ErrorBody{{Code: code, Message: msg}}},
	}
}

// Wrap превращает HandlerFunc в gin.HandlerFunc: пишет Result как есть
// (для 204 — только статус, без тела).
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		res := h(c)
		status := res.Status
		if status == 0 {
			status = http.StatusOK
		}
		if status == http.StatusNoContent {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(status, res.Body)
	}
}

// ErrorEnvelope пишет ошибку в конверте напрямую (для middleware и legacy-кода):
// {"errors":[{code,message}]}, статус — из httpx.Error.
func ErrorEnvelope(c *gin.Context, err error) {
	e := httpx.AsError(err)
	c.JSON(e.Status, EnvelopeJSON{
		Errors: []ErrorBody{{Code: e.Code, Message: e.Message}},
	})
}
