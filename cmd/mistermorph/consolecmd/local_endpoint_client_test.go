package consolecmd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/quailyquaily/mistermorph/internal/daemonruntime"
)

func TestInProcessRuntimeEndpointClientHealthMode(t *testing.T) {
	handler := daemonruntime.NewHandler(daemonruntime.RoutesOptions{
		Mode:          "console",
		AuthToken:     "dev-token",
		HealthEnabled: true,
	})
	client := newInProcessRuntimeEndpointClient(handler, "dev-token")

	mode, err := client.HealthMode(context.Background())
	if err != nil {
		t.Fatalf("HealthMode() error = %v", err)
	}
	if mode != "console" {
		t.Fatalf("HealthMode() = %q, want %q", mode, "console")
	}
}

func TestInProcessRuntimeEndpointClientProxyOverview(t *testing.T) {
	handler := daemonruntime.NewHandler(daemonruntime.RoutesOptions{
		Mode:          "console",
		AuthToken:     "dev-token",
		HealthEnabled: true,
	})
	client := newInProcessRuntimeEndpointClient(handler, "dev-token")

	status, raw, err := client.Proxy(context.Background(), http.MethodGet, "/overview", nil)
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("Proxy() status = %d, want %d (%s)", status, http.StatusOK, string(raw))
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["mode"] != "console" {
		t.Fatalf("payload.mode = %#v, want %q", payload["mode"], "console")
	}
}
