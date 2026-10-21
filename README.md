# ginserver — обёртка gin

Gin-сервер как `app.Component`: RequestID/Recovery/AccessLog/CORS, `BindJSON`, пагинация, health, единый формат ошибок httpx. Escape hatch — `Engine()`.

Зависит от [httpx](https://github.com/SaveliiYam/httpx) и [logger](https://github.com/SaveliiYam/logger).

```go
srv := ginserver.New(ginserver.Config{Addr: ":8080"}, routes, ginserver.WithLogger(log))
```

## Конверт ответов (с v0.4.0)

Обработчик возвращает готовую модель `Result{Status, Body}` через утилиты —
стабильный набор статусов: 200, 204, 401, 422, 500.

```go
func (h *Handlers) List(c *gin.Context) ginserver.Result {
	users, err := h.svc.List(c.Request.Context())
	if err != nil {
		return ginserver.Unprocessable(err) // 422, ошибка обязательна
	}
	return ginserver.Data(users) // 200 {"data": [...]}
}
```

| Утилита | Статус | Тело |
|---|---|---|
| `Data(data)` | 200 | `{"data": ...}` (пустой массив сохраняется) |
| `Empty()` | 204 | без тела |
| `Unauthorized(err)` | 401 | `{"errors":[{code,message}]}` |
| `Unprocessable(err)` | 422 | из ошибки берутся код/сообщение |
| `Internal(err)` | 500 | общее `internal error`, текст ошибки наружу не попадает |

Тело — глобальная структура `EnvelopeJSON{Data *any, Errors []ErrorBody}`
с omitempty: `"data": null` и пустой `"errors"` в ответы не попадают.
Регистрация чистых обработчиков — `ginserver.NewRouter(engine)` или реестр
yamlroutes; `Wrap` внутри библиотеки.

## Ping

По умолчанию регистрируется `GET /ping` → `200` без тела.
Отключить: `ginserver.WithoutPing()` или `Config.DisablePing: true`.

## Установка

```bash
go get github.com/SaveliiYam/ginserver@v0.4.0
```

Требуется Go ≥ 1.25.
