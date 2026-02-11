# Core-Prozessmodell

## Ziel
Minimaler Overhead. Stabil. Wartbar.

---

## 1️⃣ Prozess-Architektur

### Single-Process (MVP v0.1)

**Ein Hauptprozess: `prox-watch`**

- Daemon-Modus
- systemd-Integration
- Keine Multi-Process-Komplexität

### Threading-Modell

**Hauptthread:**
- Orchestrierung
- State-Management
- Push-Koordination

**Worker-Threads:**

1. **Log-Reader-Thread**
   - journald Streaming
   - Non-blocking
   - Queue-basiert

2. **Pattern-Matcher-Thread**
   - Pattern-Abgleich
   - Event-ID-Generierung
   - Queue-basiert

3. **Push-Thread** (optional, kann auch synchron)
   - ntfy-Versand
   - Non-blocking
   - Retry-Logik

### Kommunikation

**Thread-sichere Queues:**
- Log-Zeilen → Pattern-Matcher
- Events → State-Manager
- Push-Jobs → Push-Thread

**Shared State:**
- SQLite (thread-safe)
- In-Memory-Cache (Zähler, Cooldowns)

---

## 2️⃣ Laufzeit-Verhalten

### Start

1. Konfiguration laden
2. State-DB öffnen
3. Pattern-Definitionen laden
4. Threads starten
5. Daemon-Modus aktivieren

### Laufzeit

- Kontinuierliches Log-Streaming
- Pattern-Matching in Echtzeit
- State-Updates
- Push bei Bedarf

### Shutdown

1. Threads graceful stoppen
2. State-DB schließen
3. Ressourcen freigeben
4. Exit-Code 0

---

## 3️⃣ Restart-Verhalten

### systemd-Integration

**Unit-Definition:**
```ini
[Unit]
Description=Prox Watch Monitoring System
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/prox-watch
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Restart-Policy

**Automatisch:**
- Bei Crash (Restart=on-failure)
- Nach 10 Sekunden (RestartSec=10)
- Max. 5 Versuche / 5 Minuten

**Manuell:**
- `systemctl restart prox-watch`
- `systemctl stop prox-watch`

### State-Persistenz

- SQLite überlebt Neustarts
- Zähler bleiben erhalten
- Cooldowns bleiben erhalten
- Keine Datenverluste

---

## 4️⃣ Fehlerbehandlung

### Thread-Fehler

**Log-Reader-Thread:**
- Fehler → Loggen, Thread neu starten
- Max. 3 Neustarts → Prozess beenden

**Pattern-Matcher-Thread:**
- Fehler → Loggen, Thread neu starten
- Max. 3 Neustarts → Prozess beenden

**Push-Thread:**
- Fehler → Loggen, Retry
- Kein Thread-Neustart nötig

### Prozess-Fehler

**Kritischer Fehler:**
- State-DB nicht erreichbar → Exit 1
- Konfiguration ungültig → Exit 1
- Keine Log-Quelle → Exit 1

**Nicht-kritischer Fehler:**
- Push fehlgeschlagen → Loggen, weiter
- Pattern-Fehler → Loggen, weiter

---

## 5️⃣ Ressourcen

### Memory

- Minimaler Footprint
- Keine Memory-Leaks
- State-DB begrenzt (z. B. max. 100 MB)

### CPU

- Low-Priority (nice +10)
- Idle-Waiting (kein Busy-Loop)
- Effiziente Pattern-Matching

### Disk

- State-DB: begrenzt
- Keine Log-Speicherung
- Keine temporären Dateien

---

## 6️⃣ Monitoring des Prozesses

### Health-Check

**Intern:**
- Heartbeat in State-DB
- Alle 60 Sekunden

**Extern:**
- systemd-Status
- `systemctl status prox-watch`

### Metriken (optional, später)

- Events verarbeitet
- Patterns getroffen
- Push gesendet
- Thread-Status

---

## 7️⃣ Skalierung (später)

### Multi-Process (optional)

- Ein Prozess pro Log-Quelle
- Shared State-DB
- Koordiniertes Push

### Aktuell (MVP)

- Single-Process
- Multi-Thread
- Ausreichend für Start

---

## Status

- Prozessmodell definiert
- Single-Process, Multi-Thread
- systemd-Integration geplant
- Restart-Verhalten festgelegt
