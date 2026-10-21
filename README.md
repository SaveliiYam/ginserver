# ginserver — обёртка gin

Gin-сервер как `app.Component`: RequestID/Recovery/AccessLog/CORS, `BindJSON`, пагинация, health, единый формат ошибок httpx. Escape hatch — `Engine()`.

Зависит от [httpx](https://github.com/SaveliiYam/httpx) и [logger](https://github.com/SaveliiYam/logger).

```go
srv := ginserver.New(ginserver.Config{Addr: ":8080"}, routes, ginserver.WithLogger(log))
```

## Установка

```bash
go get github.com/SaveliiYam/ginserver@v0.1.0
```

Требуется Go ≥ 1.25. Лицензия: см. репозиторий.
