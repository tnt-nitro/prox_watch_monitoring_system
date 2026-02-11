package watcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HealthChecker prüft die Erreichbarkeit des Proxmox-Hosts.
// Phase 1: Ping + HTTPS-Check (Port 8006 oder konfigurierbar)
// Siehe docs/16_watcher_health_engine.md für vollständige Spezifikation.
type HealthChecker interface {
	Check(ctx context.Context) (Result, error)
}

// Result enthält das Ergebnis einer Health-Check-Prüfung.
// Wichtig: Kein Host, keine IP, keine Log-Inhalte.
type Result struct {
	Success   bool
	Mode      string    // "ping", "https", "ping+https"
	CheckedAt time.Time
	Latency   int       // Millisekunden (0 wenn nicht erreichbar)
	Error     string    // Leer wenn Success=true
}

// healthChecker ist die Implementierung des HealthChecker-Interfaces.
type healthChecker struct {
	host    string
	port    int
	mode    string // "ping", "https", "ping+https"
	timeout time.Duration
}

// HealthCheckerConfig enthält die Konfiguration für den HealthChecker.
type HealthCheckerConfig struct {
	Host    string
	Port    int
	Mode    string // "ping", "https", "ping+https"
	Timeout time.Duration
}

// NewHealthChecker erstellt einen neuen HealthChecker.
func NewHealthChecker(config HealthCheckerConfig) HealthChecker {
	return &healthChecker{
		host:    config.Host,
		port:    config.Port,
		mode:    config.Mode,
		timeout: config.Timeout,
	}
}

// Check führt den Health-Check aus.
// Kombiniert Ping und/oder HTTPS-Check basierend auf mode.
func (h *healthChecker) Check(ctx context.Context) (Result, error) {
	result := Result{
		Mode:      h.mode,
		CheckedAt: time.Now(),
		Latency:   0,
		Error:     "",
	}

	// Erstelle Context mit Timeout
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	var pingOK, httpsOK bool
	var latency int
	var err error

	// Ping-Check (wenn mode = "ping" oder "ping+https")
	if h.mode == "ping" || h.mode == "ping+https" {
		pingOK, latency, err = h.checkPing(checkCtx)
		if err != nil {
			result.Error = "connection_failed"
		}
	}

	// HTTPS-Check (wenn mode = "https" oder "ping+https")
	if h.mode == "https" || h.mode == "ping+https" {
		httpsOK, err = h.checkHTTPS(checkCtx)
		if err != nil {
			if result.Error == "" {
				result.Error = "tls_error"
			}
		}
	}

	// Kombinationslogik
	switch h.mode {
	case "ping":
		result.Success = pingOK
		if pingOK {
			result.Latency = latency
		}
	case "https":
		result.Success = httpsOK
	case "ping+https":
		result.Success = pingOK && httpsOK
		if pingOK {
			result.Latency = latency
		}
		if !result.Success {
			if !pingOK && !httpsOK {
				result.Error = "connection_failed"
			} else if !pingOK {
				result.Error = "connection_failed"
			} else if !httpsOK {
				result.Error = "tls_error"
			}
		}
	}

	// Timeout-Prüfung
	if checkCtx.Err() == context.DeadlineExceeded {
		result.Success = false
		result.Error = "timeout"
	}

	return result, nil
}

// checkPing führt einen Ping-Check durch.
// Verwendet TCP-Connect auf Port 22 (SSH) als Ersatz für ICMP.
// ICMP erfordert Root-Rechte, TCP-Connect ist portabler.
func (h *healthChecker) checkPing(ctx context.Context) (bool, int, error) {
	start := time.Now()
	
	// TCP-Connect auf Port 22 (SSH) als Ping-Ersatz
	// Alternativ könnte Port 80 (HTTP) verwendet werden
	pingPort := 22
	if h.port > 0 {
		// Verwende konfigurierten Port als Fallback
		pingPort = h.port
	}

	dialer := &net.Dialer{
		Timeout: h.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", h.host, pingPort))
	if err != nil {
		return false, 0, err
	}
	defer conn.Close()

	latency := int(time.Since(start).Milliseconds())
	return true, latency, nil
}

// checkHTTPS führt einen HTTPS-Check durch.
// Prüft nur Verbindungsaufbau, keine Zertifikatsprüfung im MVP.
func (h *healthChecker) checkHTTPS(ctx context.Context) (bool, error) {
	client := &http.Client{
		Timeout: h.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // MVP: Keine Zertifikatsprüfung
			},
		},
	}

	url := fmt.Sprintf("https://%s:%d", h.host, h.port)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Status-Code irrelevant, nur Verbindungsaufbau zählt
	return true, nil
}
