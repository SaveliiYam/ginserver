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
| `InvalidToken(err)` | 401 | code `invalid_token` (битый/просроченный JWT) |
| `Validation(msg)` | 422 | code `validation` (входные данные) |
| `Unprocessable(err)` | 422 | из ошибки берутся код/сообщение |
| `Internal(err)` | 500 | общее `internal error`, текст ошибки наружу не попадает |

Тело — глобальная структура `EnvelopeJSON{Data *any, Errors []ErrorBody}`
с omitempty: `"data": null` и пустой `"errors"` в ответы не попадают.
Регистрация чистых обработчиков — `ginserver.NewRouter(engine)` или реестр
yamlroutes; `Wrap` внутри библиотеки.

## Логирование запросов (с v0.5.0)

`RequestLogger` — обязательная middleware: ставится в `New` всегда,
отключить нельзя. Каждый запрос — одна запись после завершения:
поля `method, path, route, status, duration_ms, bytes, client_ip, request_id`
(+ `errors`, если есть). Уровень по статусу: 5xx — error, 4xx — warn,
иначе info. Логгер по умолчанию — slog JSON в stdout, свой — `WithLogger(...)`
(например, `logger.Nop{}`, если логи ведёт внешняя система).
`AccessLog` оставлен deprecated-алиасом.

## Ping

По умолчанию регистрируется `GET /ping` → `200` без тела.
Отключить: `ginserver.WithoutPing()` или `Config.DisablePing: true`.

## HTTP-модуль из JSON (с v0.7.0)

Секция `http.json` + `config.Apply` без копипасты в каждом сервисе:

```go
httpCfg := ginserver.DefaultModuleConfig()
_, _ = config.LoadSection(boot, "http", &httpCfg)
transport := ginserver.NewModule()
_ = config.Apply(ctx, transport, httpCfg)

srv := ginserver.New(transport.ServerConfig(), routes)
_ = transport.Config().RoutesFile // yamlroutes
```

Поля: `addr`, `gin_mode`, `routes_file`, `routes_mode` (`yaml`|`code`).

## Basic Auth (с v0.6.0)

Учётки — JSON-файл `{"login1":"password1","login2":"password2"}`
или `ginserver.Accounts` из конфига. Пустой map — middleware no-op.

```go
accounts, err := ginserver.LoadAccountsJSON("config/dev/jwt/basic_auth.json")
// или Decode в Accounts из секции конфига

r.Use(ginserver.BasicAuth(accounts,
    ginserver.WithRealm("jwt"),
    ginserver.WithSkipPaths("/healthz", "/ping"), // так и по умолчанию
))
```

При отказе — `401` в конверте `{"errors":[...]}` и заголовок `WWW-Authenticate`.

## Установка

```bash
go get github.com/SaveliiYam/ginserver@v0.7.0
```

Требуется Go ≥ 1.25.
