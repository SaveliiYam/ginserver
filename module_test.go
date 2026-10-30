package ginserver_test

import (
	"context"
	"testing"

	"github.com/SaveliiYam/ginserver"
)

func TestModuleConfigureDefaultsAndValidation(t *testing.T) {
	m := ginserver.NewModule()
	if err := m.Configure(context.Background(), ginserver.ModuleConfig{GinMode: "debug"}); err != nil {
		t.Fatal(err)
	}
	cfg := m.Config()
	if cfg.Addr != ":8080" || cfg.RoutesFile != "routes.yaml" || cfg.RoutesMode != "yaml" {
		t.Fatalf("%+v", cfg)
	}
	sc := m.ServerConfig()
	if sc.Addr != ":8080" || sc.Mode != "debug" {
		t.Fatalf("%+v", sc)
	}

	if err := m.Configure(context.Background(), ginserver.ModuleConfig{GinMode: "nope"}); err == nil {
		t.Fatal("expected gin_mode error")
	}
	if err := m.Configure(context.Background(), ginserver.ModuleConfig{GinMode: "release", RoutesMode: "x"}); err == nil {
		t.Fatal("expected routes_mode error")
	}
}
