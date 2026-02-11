# Externer Wächter (Raspberry Pi)

## Status

**Phase 1 - Implementiert** ✅

## Übersicht

Der externe Wächter ist ein Out-of-Band-Monitoring-System, das unabhängig vom Proxmox-Host läuft (z.B. auf einem Raspberry Pi). Er erkennt Totalausfälle, wenn Proxmox selbst nicht mehr antwortet.

## Architektur

Der Watcher ist Teil des Monorepos und nutzt shared packages:
- `cmd/watcher/` - Binary für externen Wächter
- `internal/watcher/` - Watcher-spezifische Komponenten
- Shared: `internal/push`, `internal/rules`, `internal/config`

## MVP-Funktionen (Phase 1)

- ✅ Erreichbarkeitsprüfung (Ping + HTTPS)
- ✅ Fehlzähler (WARN ≥3, CRIT ≥10)
- ✅ Push an ntfy (gleiche Topics wie Core)
- ✅ Lokale Signale (LED + Beeper)

## Hardware (geplant)

- **Raspberry Pi** (beliebiges Modell)
- **GPIO** - LED (Grün/Gelb/Rot) + Beeper
- **Netzwerk** - Erreichbarkeit zu Proxmox-Host

## Implementierung

Siehe [cmd/watcher/README.md](../cmd/watcher/README.md) für Details.

## Power-Cycle (optional, später)

Power-Cycle-Funktionalität ist **nicht** Teil von Phase 1. Design folgt später (gesperrt, manuelle Freigabe erforderlich).
