package watcher

import (
	"time"
)

// isDayTime prüft, ob die aktuelle Zeit im Tag-Zeitfenster liegt.
// Phase 1.5: Implementierung mit time.Now().
// Format: windowStart und windowEnd im Format "HH:MM" (lokale Zeit).
// Vergleich nur lokal (kein TZ-Handling nötig).
// Unterstützt Nacht-Fenster (z.B. 22:00-06:00).
func isDayTime(windowStart, windowEnd string) bool {
	// Parse Zeitfenster
	startTime, err := parseTimeHM(windowStart)
	if err != nil {
		// Fehler beim Parsen → immer true (Fallback)
		return true
	}

	endTime, err := parseTimeHM(windowEnd)
	if err != nil {
		// Fehler beim Parsen → immer true (Fallback)
		return true
	}

	// Aktuelle lokale Zeit
	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.Local)

	// Prüfe, ob aktuelle Zeit im Fenster liegt
	if startTime.Before(endTime) || startTime.Equal(endTime) {
		// Normalfall: Start ≤ End (z.B. 08:00 - 22:00)
		// Inklusive Grenzen: >= Start && <= End
		return (currentTime.After(startTime) || currentTime.Equal(startTime)) &&
			(currentTime.Before(endTime) || currentTime.Equal(endTime))
	} else {
		// Über-Mitternacht: Start > End (z.B. 22:00 - 06:00)
		// Zeit liegt im Fenster, wenn sie >= Start ODER <= End
		return currentTime.After(startTime) || currentTime.Equal(startTime) ||
			currentTime.Before(endTime) || currentTime.Equal(endTime)
	}
}

// parseTimeHM parst eine Zeit im Format "HH:MM" (lokale Zeit).
func parseTimeHM(v string) (time.Time, error) {
	return time.Parse("15:04", v)
}
