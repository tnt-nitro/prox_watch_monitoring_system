package watcher

import (
	"context"
	"prox-watch/internal/push"
	"prox-watch/internal/rules"
	"time"
)

// PushService orchestriert Push-Benachrichtigungen für den Watcher.
// Nutzt den shared Push-Adapter aus internal/push.
// Siehe docs/17_watcher_counter_severity.md für vollständige Spezifikation.
type PushService struct {
	adapter push.Adapter
	enabled bool
}

// PushServiceConfig enthält die Konfiguration für den PushService.
type PushServiceConfig struct {
	Adapter push.Adapter
	Enabled bool
}

// NewPushService erstellt einen neuen PushService.
func NewPushService(config PushServiceConfig) *PushService {
	return &PushService{
		adapter: config.Adapter,
		enabled: config.Enabled,
	}
}

// SendIfEscalation sendet eine Push-Benachrichtigung nur bei Eskalation.
// Push nur wenn: newSeverity > currentSeverity
//
// Event-ID: external.availability.down
// Topics: identisch zum Core (prox-watch-warn, prox-watch-crit)
//
// Fehlerverhalten: Nicht blockierend, kein Retry
func (p *PushService) SendIfEscalation(ctx context.Context, currentSeverity, newSeverity rules.Severity) error {
	// Push deaktiviert
	if !p.enabled {
		return nil
	}

	// Push nur bei Eskalation (newSeverity > currentSeverity)
	if !ShouldPush(currentSeverity, newSeverity) {
		return nil
	}

	// Push nur für WARN und CRIT (INFO hat kein Push)
	if newSeverity == rules.SeverityInfo {
		return nil
	}

	// Hole Topic für Severity
	topic := p.adapter.GetTopic(newSeverity)

	// Erstelle Push-Message
	msg := push.Message{
		EventID:   "external.availability.down",
		Severity: newSeverity,
		Timestamp: time.Now().UTC(),
	}

	// Sende Push (nicht blockierend, Fehler werden ignoriert)
	err := p.adapter.Send(ctx, topic, msg)
	if err != nil {
		// Push-Fehler werden nicht eskaliert, nur geloggt (optional)
		// Kein Retry, kein Blocking
		return err
	}

	return nil
}

// IsEnabled gibt zurück, ob Push aktiviert ist.
func (p *PushService) IsEnabled() bool {
	return p.enabled
}
