package ginserver

import (
	"context"
	"fmt"
)

// ModuleConfig — JSON-секция http.json (config.LoadSection).
// Отдельно от Config сервера: здесь ещё метаданные маршрутов.
type ModuleConfig struct {
	Addr       string `json:"addr"`
	GinMode    string `json:"gin_mode"`
	RoutesFile string `json:"routes_file"`
	RoutesMode string `json:"routes_mode"` // yaml | code
}

// DefaultModuleConfig — дефолты, если секция http выключена.
func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		Addr:       ":8080",
		GinMode:    "release",
		RoutesFile: "routes.yaml",
		RoutesMode: "yaml",
	}
}

// Module — HTTP-транспорт как настраиваемый слой
// (структурно config.Configurable[ModuleConfig]).
type Module struct {
	cfg ModuleConfig
}

// NewModule создаёт Module с DefaultModuleConfig.
func NewModule() *Module {
	return &Module{cfg: DefaultModuleConfig()}
}

// Configure применяет http.json.
func (m *Module) Configure(_ context.Context, cfg ModuleConfig) error {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.RoutesFile == "" {
		cfg.RoutesFile = "routes.yaml"
	}
	if cfg.RoutesMode == "" {
		cfg.RoutesMode = "yaml"
	}
	switch cfg.GinMode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("ginserver: gin_mode must be debug|release|test, got %q", cfg.GinMode)
	}
	switch cfg.RoutesMode {
	case "yaml", "code":
	default:
		return fmt.Errorf("ginserver: routes_mode must be yaml|code, got %q", cfg.RoutesMode)
	}
	m.cfg = cfg
	return nil
}

// Config возвращает применённый ModuleConfig.
func (m *Module) Config() ModuleConfig { return m.cfg }

// ServerConfig — конфиг для New.
func (m *Module) ServerConfig() Config {
	return Config{Addr: m.cfg.Addr, Mode: m.cfg.GinMode}
}
