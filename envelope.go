package ginserver

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/httpx"
)

// Envelope — единый формат любого ответа с телом:
//
//	успех: {"data": {...} | [...]}
//	ошибка: {"errors": [{"code": "...", "message": "..."}]}
//
// Оба поля с omitempty-семантикой: пустое поле в JSON не попадает,
// поэтому "data": null и "errors": [] в ответах не встречаются.
// Пустой, но не nil, срез данных ([]) при этом сохраняется.
//
// Если данных нет и ошибки нет — тело не отправляется (204, см. Wrap).
type Envelope struct {
	Data   any         `json:"data,omitempty"`
	Errors []ErrorBody `json:"errors,omitempty"`
}

// ErrorBody — одна ошибка в конверте.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MarshalJSON опускает nil data и пустые errors, но сохраняет пустой срез data.
func (e Envelope) MarshalJSON() ([]byte, error) {
	type envelopeJSON struct {
		Data   *any        `json:"data,omitempty"`
		Errors []ErrorBody `json:"errors,omitempty"`
	}
	out := envelopeJSON{Errors: e.Errors}
	if e.Data != nil {
		out.Data = &e.Data
	}
	return json.Marshal(out)
}

// HandlerFunc — обработчик, возвращающий данные и ошибку, ничего не пишет сам.
type HandlerFunc func(c *gin.Context) (any, error)

// Wrap превращает HandlerFunc в gin.HandlerFunc (статус 200 при успехе).
//
//	data != nil, err == nil -> 200 {"data": data}
//	data == nil, err == nil -> 204 No Content, без тела
//	err != nil               -> статус из httpx.Error, {"errors": [{...}]}
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return WrapStatus(http.StatusOK, h)
}

// WrapStatus — то же, но с заданным статусом успеха (например, 201 для создания).
func WrapStatus(status int, h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c)
		if err != nil {
			ErrorEnvelope(c, err)
			return
		}
		if data == nil {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(status, Envelope{Data: data})
	}
}

// ErrorEnvelope пишет ошибку в конверте: {"errors":[{code,message}]}, без data.
func ErrorEnvelope(c *gin.Context, err error) {
	e := httpx.AsError(err)
	c.JSON(e.Status, Envelope{
		Errors: []ErrorBody{{Code: e.Code, Message: e.Message}},
	})
}
