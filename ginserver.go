// Package ginserver — удобная обёртка над gin за интерфейсами httpx/app.
//
// Замена фреймворка: Server реализует httpx.Server и app.Component
// (структурно). Если завтра нужен chi/echo — пишется соседний адаптер
// с тем же контрактом, роуты уровня сервиса переезжают, main не меняется
// кроме одной строки конструктора.
package ginserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/SaveliiYam/httpx"
	"github.com/SaveliiYam/logger"
)

// Config — настройки HTTP-сервера.
type Config struct {
	Addr            string        `env:"ADDR" envDefault:":8080"`
	Mode            string        `env:"MODE" envDefault:"release"` // debug|release|test
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	TrustedProxies  []string      `env:"TRUSTED_PROXIES"`
	// DisablePing отключает встроенный GET /ping (200 без тела), включён по умолчанию.
	DisablePing bool `env:"DISABLE_PING"`
}

func (c *Config) withDefaults() {
	if c.Addr == "" {
		c.Addr = ":8080"
	}
	if c.Mode == "" {
		c.Mode = gin.ReleaseMode
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 10 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 30 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 60 * time.Second
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 10 * time.Second
	}
}

// Server — gin + http.Server как app.Component.
type Server struct {
	cfg    Config
	engine *gin.Engine
	srv    *http.Server
	log    logger.Logger
	extra  []gin.HandlerFunc
}

// Option настраивает Server.
type Option func(*Server)

// WithLogger задаёт логгер middleware и сервера.
func WithLogger(l logger.Logger) Option {
	return func(s *Server) {
		if l != nil {
			s.log = l
		}
	}
}

// WithMiddleware добавляет gin-middleware (после встроенных RequestID/Recovery/AccessLog).
func WithMiddleware(mw ...gin.HandlerFunc) Option {
	return func(s *Server) { s.extra = append(s.extra, mw...) }
}

// WithoutPing отключает встроенный GET /ping.
func WithoutPing() Option {
	return func(s *Server) { s.cfg.DisablePing = true }
}

// PingPath — путь встроенного ping.
const PingPath = "/ping"

// PingHandler — 200 без тела. Публичный, чтобы можно было повесить и вручную.
func PingHandler(c *gin.Context) { c.Status(http.StatusOK) }

// Routes — функция регистрации маршрутов.
type Routes func(r gin.IRouter)

// New создаёт Server. routes может быть nil (добавить позже через Engine/Group).
// Встроенные middleware: RequestID, Recovery, AccessLog. Gin Recovery не ставится.
func New(cfg Config, routes Routes, opts ...Option) *Server {
	cfg.withDefaults()
	gin.SetMode(cfg.Mode)
	e := gin.New()
	_ = e.SetTrustedProxies(cfg.TrustedProxies)

	s := &Server{cfg: cfg, engine: e, log: logger.Nop{}}
	for _, o := range opts {
		o(s)
	}
	// Порядок middleware: встроенные -> пользовательские -> хендлеры.
	e.Use(RequestID(), Recovery(s.log), AccessLog(s.log))
	if len(s.extra) > 0 {
		e.Use(s.extra...)
	}
	if !s.cfg.DisablePing {
		e.GET(PingPath, PingHandler) // по умолчанию: GET /ping -> 200 без тела
	}
	if routes != nil {
		routes(e)
	}
	s.srv = &http.Server{
		Addr: cfg.Addr, Handler: e,
		ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout,
	}
	return s
}

// Engine — нативный *gin.Engine (escape hatch: весь gin доступен).
func (s *Server) Engine() *gin.Engine { return s.engine }

// Group — шорткат для группы маршрутов.
func (s *Server) Group(prefix string, mw ...gin.HandlerFunc) *gin.RouterGroup {
	return s.engine.Group(prefix, mw...)
}

// Name — app.Component.
func (s *Server) Name() string { return "http" }

// Addr фактический адрес из конфига.
func (s *Server) Addr() string { return s.cfg.Addr }

// Start блокируется до отмены ctx или ошибки ListenAndServe. app.Component.
func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("ginserver: listen %s: %w", s.cfg.Addr, err)
	}
	s.log.Info(ctx, "http server listening", "addr", ln.Addr().String())
	errCh := make(chan error, 1)
	go func() { errCh <- s.srv.Serve(ln) }()
	select {
	case <-ctx.Done():
		return s.Stop(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// Stop — graceful shutdown. app.Stopper.
func (s *Server) Stop(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.srv.Shutdown(c); err != nil {
		return fmt.Errorf("ginserver: shutdown: %w", err)
	}
	return nil
}

// Компиляторные проверки контрактов.
var (
	_ httpx.Server = (*Server)(nil)
)
