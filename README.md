# ginserver — обёртка gin

Gin-сервер как `app.Component`: RequestID/Recovery/AccessLog/CORS, `BindJSON`, пагинация, health, единый формат ошибок httpx. Escape hatch — `Engine()`.

Зависит от [httpx](https://github.com/SaveliiYam/httpx) и [logger](https://github.com/SaveliiYam/logger).

```go
srv := ginserver.New(ginserver.Config{Addr: ":8080"}, routes, ginserver.WithLogger(log))
```

## Конверт ответов (с v0.2.0)

Обработчики возвращают `(data, error)`, обёртка формирует единый ответ:

```go
e.GET("/users", ginserver.Wrap(handlers.List))
e.POST("/users", ginserver.WrapStatus(http.StatusCreated, handlers.Create))
```

| Результат | Ответ |
|---|---|
| `(объект, nil)` | `200 {"data": {...}}` |
| `(массив, nil)` | `200 {"data": [...]}` (пустой массив сохраняется) |
| `(nil, ошибка)` | статус из ошибки, `{"errors": [{"code", "message"}]}` |
| `(nil, nil)` | `204 No Content` без тела |

Поля — с omitempty-семантикой: `"data": null` и пустой `"errors": []`
в ответы не попадают.

## Ping

По умолчанию регистрируется `GET /ping` → `200` без тела.
Отключить: `ginserver.WithoutPing()` или `Config.DisablePing: true`.

## Установка

```bash
go get github.com/SaveliiYam/ginserver@v0.3.0
```

Требуется Go ≥ 1.25.
