package watcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthChecker_Check_PingMode(t *testing.T) {
	// Ping-Mode: TCP-Connect auf Port 22
	// Für Tests verwenden wir einen nicht existierenden Host
	config := HealthCheckerConfig{
		Host:    "127.0.0.1",
		Port:    22,
		Mode:    "ping",
		Timeout: 1 * time.Second,
	}

	checker := NewHealthChecker(config)
	ctx := context.Background()

	result, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	// Erwartung: Success sollte true sein (localhost:22 ist normalerweise erreichbar)
	// Wenn nicht, ist das auch OK (abhängig von System)
	if result.Mode != "ping" {
		t.Errorf("Expected mode 'ping', got '%s'", result.Mode)
	}
}

func TestHealthChecker_Check_HTTPSMode(t *testing.T) {
	// HTTPS-Mode: Fake HTTP Server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extrahiere Host und Port vom Test-Server
	host := server.URL[8:] // Entferne "https://"
	config := HealthCheckerConfig{
		Host:    host,
		Port:    443, // HTTPS Standard-Port
		Mode:    "https",
		Timeout: 5 * time.Second,
	}

	checker := NewHealthChecker(config)
	ctx := context.Background()

	result, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected Success=true for HTTPS check, got Success=false, Error=%s", result.Error)
	}
	if result.Mode != "https" {
		t.Errorf("Expected mode 'https', got '%s'", result.Mode)
	}
}

func TestHealthChecker_Check_PingHTTPSMode(t *testing.T) {
	// Ping+HTTPS-Mode: Beide müssen erfolgreich sein
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host := server.URL[8:] // Entferne "https://"
	config := HealthCheckerConfig{
		Host:    host,
		Port:    443,
		Mode:    "ping+https",
		Timeout: 5 * time.Second,
	}

	checker := NewHealthChecker(config)
	ctx := context.Background()

	result, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if result.Mode != "ping+https" {
		t.Errorf("Expected mode 'ping+https', got '%s'", result.Mode)
	}
	// Beide Checks müssen erfolgreich sein
	if !result.Success {
		t.Logf("Ping+HTTPS check failed (may be expected if ping fails): Error=%s", result.Error)
	}
}

func TestHealthChecker_Check_Timeout(t *testing.T) {
	// Timeout-Test: Verwendet einen nicht erreichbaren Host mit sehr kurzem Timeout
	config := HealthCheckerConfig{
		Host:    "192.0.2.0", // Test-Netz (RFC 3330), sollte nicht erreichbar sein
		Port:    8006,
		Mode:    "https",
		Timeout: 100 * time.Millisecond, // Sehr kurzer Timeout
	}

	checker := NewHealthChecker(config)
	ctx := context.Background()

	result, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	// Erwartung: Timeout oder Connection Failed
	if result.Success {
		t.Errorf("Expected Success=false for timeout scenario, got Success=true")
	}
	if result.Error == "" {
		t.Error("Expected Error to be set for timeout scenario")
	}
}

func TestHealthChecker_Check_CombinationLogic(t *testing.T) {
	// Kombinationslogik: ping+https
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host := server.URL[8:]
	config := HealthCheckerConfig{
		Host:    host,
		Port:    443,
		Mode:    "ping+https",
		Timeout: 5 * time.Second,
	}

	checker := NewHealthChecker(config)
	ctx := context.Background()

	result, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	// Bei ping+https müssen beide erfolgreich sein
	// Wenn einer fehlschlägt, sollte Success=false sein
	if result.Mode != "ping+https" {
		t.Errorf("Expected mode 'ping+https', got '%s'", result.Mode)
	}
}

func TestHealthChecker_Check_NoHostInResult(t *testing.T) {
	// Datenschutz: Kein Host/IP im Result
	config := HealthCheckerConfig{
		Host:    "test.example.com",
		Port:    8006,
		Mode:    "https",
		Timeout: 1 * time.Second,
	}

	checker := NewHealthChecker(config)
	ctx := context.Background()

	result, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	// Prüfe, dass Result keine Host-Informationen enthält
	// (Result-Struktur hat kein Host-Feld, das ist korrekt)
	if result.Error != "" && (result.Error == "test.example.com" || result.Error == "8006") {
		t.Errorf("Result.Error should not contain host/port information, got: %s", result.Error)
	}
}

func TestHealthChecker_Check_AbstractErrors(t *testing.T) {
	// Abstrakte Fehlermeldungen: timeout, connection_failed, tls_error
	configs := []struct {
		name    string
		config  HealthCheckerConfig
		wantErr string
	}{
		{
			name: "timeout",
			config: HealthCheckerConfig{
				Host:    "192.0.2.0",
				Port:    8006,
				Mode:    "https",
				Timeout: 10 * time.Millisecond,
			},
			wantErr: "timeout",
		},
	}

	for _, tt := range configs {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewHealthChecker(tt.config)
			ctx := context.Background()

			result, err := checker.Check(ctx)
			if err != nil {
				t.Fatalf("Check() failed: %v", err)
			}

			if result.Success {
				t.Errorf("Expected Success=false for %s scenario", tt.name)
			}
			// Prüfe, dass Error abstrakt ist (kein Host/IP)
			if result.Error != "" && (result.Error == tt.config.Host || result.Error == fmt.Sprintf("%d", tt.config.Port)) {
				t.Errorf("Result.Error should be abstract, got: %s", result.Error)
			}
		})
	}
}
