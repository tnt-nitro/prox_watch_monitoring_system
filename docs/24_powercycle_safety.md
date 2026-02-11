# Power-Cycle Sicherheitsdokumentation

**⚠️ KRITISCH: Diese Funktion steuert physische Hardware und kann zu Datenverlust oder Systemausfällen führen.**

## Übersicht

Die Power-Cycle-Funktionalität ermöglicht es dem externen Wächter, bei kritischen Systemausfällen automatisch einen Power-Cycle (Strom trennen und wieder einschalten) des überwachten Proxmox-Hosts durchzuführen.

**Diese Funktion ist standardmäßig DEAKTIVIERT** und erfordert explizite manuelle Aktivierung.

## Sicherheitsmechanismen

### 1. Mehrschichtige Schutzlogik

Die Power-Cycle-Funktionalität ist durch mehrere Sicherheitsebenen geschützt:

#### 🔒 Severity-Edge-Trigger
- Power-Cycle wird **nur** ausgelöst, wenn die Severity von **< CRIT → CRIT** wechselt
- **NICHT** bei dauerhaftem CRIT (verhindert Neustart-Loops)
- **NICHT** nach Restart ohne echte Eskalation

#### 🔒 ARM-Datei (Manuelle Freigabe)
- Power-Cycle erfordert eine ARM-Datei (`/var/lib/prox-watch/arm_powercycle` per Default)
- Die Datei muss **manuell** erstellt werden, um einen Power-Cycle zu erlauben
- Nach erfolgreichem Attempt wird die ARM-Datei **automatisch entfernt**
- Bei Fehler bleibt die ARM-Datei erhalten (für Diagnose)

#### 🔒 Max Attempts
- Konfigurierbare maximale Anzahl von Versuchen (`max_attempts`, Default: 1)
- Nach Erreichen des Limits wird kein weiterer Attempt durchgeführt

#### 🔒 Retry-Cooldown
- Konfigurierbarer Cooldown zwischen Versuchen (`retry_after_seconds`, Default: 900 = 15 Minuten)
- Verhindert wiederholte Versuche innerhalb des Cooldown-Fensters

#### 🔒 Min Downtime
- Konfigurierbare Mindest-Downtime (`min_downtime_seconds`, Default: 15 Sekunden)
- Garantiert, dass der Host ausreichend Zeit hat, vollständig herunterzufahren

### 2. Kill-Switch

**Schnellste Deaktivierung:**
```yaml
powercycle:
  enabled: false
```

Nach Änderung der Konfiguration muss der Watcher neu gestartet werden:
```bash
sudo systemctl restart prox-watch-watcher
```

## Konfiguration

### Minimal-Konfiguration

```yaml
powercycle:
  enabled: true
  gpio_pin: 24                    # BCM Pin-Nummer
  relay_active_high: false       # true = Active HIGH, false = Active LOW
  relay_mode: "cut_power_on_inactive"  # MUSS gesetzt werden!
  max_attempts: 1
  min_downtime_seconds: 15
  retry_after_seconds: 900
  require_manual_arm: true
  arm_file_path: "/var/lib/prox-watch/arm_powercycle"
```

### Konfigurationsparameter

| Parameter | Typ | Default | Beschreibung |
|-----------|-----|---------|--------------|
| `enabled` | bool | `false` | Aktiviert/deaktiviert Power-Cycle |
| `gpio_pin` | int | `24` | BCM Pin-Nummer für Relais |
| `relay_active_high` | bool | `false` | `true` = Active HIGH, `false` = Active LOW |
| `relay_mode` | string | - | **MUSS gesetzt werden!** Siehe unten |
| `max_attempts` | int | `1` | Maximale Anzahl Versuche |
| `min_downtime_seconds` | int | `15` | Mindest-Downtime in Sekunden |
| `retry_after_seconds` | int | `900` | Cooldown zwischen Versuchen (Sekunden) |
| `require_manual_arm` | bool | `true` | Erfordert ARM-Datei |
| `arm_file_path` | string | `/var/lib/prox-watch/arm_powercycle` | Pfad zur ARM-Datei |

### `relay_mode` (Pflicht!)

**⚠️ KRITISCH: Dieser Parameter MUSS korrekt gesetzt werden, basierend auf Ihrer Hardware-Verdrahtung.**

#### `cut_power_on_active`
- **Typisch für NO (Normally Open) Relais**
- Relais trennt Strom, wenn **aktiv** (LOW bei Active LOW, HIGH bei Active HIGH)
- **Verdrahtung:** Relais schließt bei Aktivierung → Strom getrennt

#### `cut_power_on_inactive`
- **Typisch für NC (Normally Closed) Relais**
- Relais trennt Strom, wenn **inaktiv** (HIGH bei Active LOW, LOW bei Active HIGH)
- **Verdrahtung:** Relais öffnet bei Deaktivierung → Strom getrennt

**Wie bestimme ich meinen `relay_mode`?**

1. **Prüfe Ihre Relais-Dokumentation:** Ist es NO oder NC?
2. **Test ohne Last:** Schließen Sie das Relais manuell und prüfen Sie, ob Strom fließt oder getrennt ist
3. **Bei Unsicherheit:** Verwenden Sie `cut_power_on_inactive` (häufiger bei Optokoppler-Modulen)

## Wiring-Hinweise

### Hardware-Anforderungen

- **Raspberry Pi** mit GPIO-Zugriff
- **Relais-Modul** (z.B. 1-Kanal Optokoppler-Relais)
- **Verdrahtung:** Relais zwischen Netzteil und Proxmox-Host

### Pin-Layout (BCM-Nummerierung)

```
Raspberry Pi GPIO:
  Pin 24 (BCM) → Relais-Modul IN
  GND → Relais-Modul GND
  3.3V → Relais-Modul VCC (falls benötigt)
```

### Verdrahtung

**⚠️ WICHTIG: Prüfen Sie die Verdrahtung VOR der Aktivierung!**

1. **Relais-Modul an Raspberry Pi anschließen:**
   - IN → GPIO Pin (z.B. BCM 24)
   - GND → GND
   - VCC → 3.3V (falls benötigt)

2. **Relais zwischen Netzteil und Host:**
   - **NO (Normally Open):** Relais schließt bei Aktivierung → Strom fließt
   - **NC (Normally Closed):** Relais öffnet bei Deaktivierung → Strom getrennt

3. **Test ohne Last:**
   - Manuelles Schließen/Öffnen des Relais
   - Prüfen Sie, ob Strom getrennt wird
   - Bestimmen Sie `relay_mode` basierend auf dem Verhalten

### Sicherheitshinweise

- **Isolierung:** Verwenden Sie ein Optokoppler-Relais für galvanische Trennung
- **Last:** Prüfen Sie, ob das Relais für die Last des Proxmox-Hosts ausgelegt ist
- **Fuses:** Verwenden Sie Sicherungen für zusätzlichen Schutz
- **Test:** Testen Sie die Schaltung **ohne** echten Host (z.B. mit LED)

## ARM-Prozess

### Aktivierung (ARM-Datei erstellen)

Um einen Power-Cycle zu erlauben, muss die ARM-Datei erstellt werden:

```bash
# Als root oder mit sudo
sudo touch /var/lib/prox-watch/arm_powercycle
sudo chmod 600 /var/lib/prox-watch/arm_powercycle
```

**WICHTIG:**
- Die ARM-Datei wird **automatisch entfernt** nach erfolgreichem Attempt
- Bei Fehler bleibt die ARM-Datei erhalten (für Diagnose)
- Für jeden neuen Attempt muss die ARM-Datei **neu erstellt** werden

### Deaktivierung (ARM-Datei entfernen)

```bash
# Als root oder mit sudo
sudo rm /var/lib/prox-watch/arm_powercycle
```

**WICHTIG:**
- Entfernen der ARM-Datei verhindert **sofort** weitere Attempts
- Bestehende Attempts werden nicht abgebrochen (laufen zu Ende)

### Automatische Deaktivierung

Die ARM-Datei wird automatisch entfernt:
- ✅ Nach erfolgreichem Power-Cycle-Versuch
- ❌ **NICHT** bei Fehler (bleibt für Diagnose)

## Fehlerbehandlung

### Häufige Fehler

#### "powercycle: not armed (arm file missing)"
- **Ursache:** ARM-Datei fehlt
- **Lösung:** ARM-Datei erstellen (siehe oben)

#### "powercycle: max attempts (X) reached"
- **Ursache:** Maximale Anzahl Versuche erreicht
- **Lösung:** `max_attempts` in Konfiguration erhöhen (mit Vorsicht!)

#### "powercycle: retry cooldown active"
- **Ursache:** Retry-Cooldown noch aktiv
- **Lösung:** Warten bis Cooldown abgelaufen ist oder `retry_after_seconds` reduzieren

#### "powercycle: relay_mode must be set"
- **Ursache:** `relay_mode` fehlt in Konfiguration
- **Lösung:** `relay_mode` setzen (siehe oben)

#### "powercycle: GPIO operation failed"
- **Ursache:** Fehler bei GPIO-Schaltung
- **Lösung:** 
  - Prüfen Sie die Pin-Verdrahtung
  - Prüfen Sie, ob der Pin bereits verwendet wird
  - Prüfen Sie die GPIO-Berechtigungen

### Diagnose

Bei Fehler bleibt die ARM-Datei erhalten. Prüfen Sie:

1. **State-Datenbank:**
   ```bash
   sqlite3 /var/lib/prox-watch-watcher/watcher_state.db "SELECT * FROM watcher_state;"
   ```

2. **Watcher-Logs:**
   ```bash
   sudo journalctl -u prox-watch-watcher -n 50
   ```

3. **GPIO-Status:**
   ```bash
   # Prüfen Sie, ob der Pin korrekt konfiguriert ist
   gpio readall  # Falls verfügbar
   ```

## Risiken und Warnungen

### ⚠️ KRITISCHE RISIKEN

1. **Datenverlust:**
   - Power-Cycle kann zu Datenverlust führen, wenn der Host nicht sauber herunterfährt
   - **Empfehlung:** Verwenden Sie nur bei absoluten Notfällen

2. **Hardware-Schäden:**
   - Wiederholte Power-Cycles können Hardware-Schäden verursachen
   - **Empfehlung:** Setzen Sie `max_attempts` auf 1 (Default)

3. **Neustart-Loops:**
   - Bei falscher Konfiguration kann ein Neustart-Loop entstehen
   - **Schutz:** Severity-Edge-Trigger verhindert dies

4. **Falsche Verdrahtung:**
   - Falsche `relay_mode`-Konfiguration kann zu unerwartetem Verhalten führen
   - **Empfehlung:** Testen Sie die Schaltung **ohne** echten Host

### ⚠️ WARNUNGEN

- **Nicht für Produktionsumgebungen ohne Tests:** Testen Sie die Funktionalität in einer sicheren Umgebung
- **Backup:** Stellen Sie sicher, dass Backups vorhanden sind
- **Monitoring:** Überwachen Sie die Power-Cycle-Aktivität (Logs, State-DB)
- **Wartung:** Prüfen Sie regelmäßig die Hardware-Verdrahtung

## Best Practices

1. **Minimale Konfiguration:**
   - `max_attempts: 1` (Default)
   - `min_downtime_seconds: 15` (Default)
   - `retry_after_seconds: 900` (15 Minuten, Default)

2. **Test-Prozess:**
   - Testen Sie die Schaltung **ohne** echten Host
   - Testen Sie mit LED oder Multimeter
   - Prüfen Sie die Pin-Sequenz (Power OFF → Warte → Power ON)

3. **Monitoring:**
   - Überwachen Sie `PowerAttempts` in der State-DB
   - Überwachen Sie die ARM-Datei (wird entfernt nach Attempt)
   - Überwachen Sie Watcher-Logs

4. **Wartung:**
   - Regelmäßige Prüfung der Hardware-Verdrahtung
   - Regelmäßige Prüfung der Konfiguration
   - Regelmäßige Prüfung der State-DB

## Deaktivierung

### Temporäre Deaktivierung

```bash
# ARM-Datei entfernen
sudo rm /var/lib/prox-watch/arm_powercycle
```

### Permanente Deaktivierung

```yaml
powercycle:
  enabled: false
```

Nach Änderung:
```bash
sudo systemctl restart prox-watch-watcher
```

## Support

Bei Problemen oder Fragen:
1. Prüfen Sie diese Dokumentation
2. Prüfen Sie die Watcher-Logs
3. Prüfen Sie die State-Datenbank
4. Erstellen Sie ein Issue im Repository (mit anonymisierten Logs)
