# GitHub-Upload Anleitung

## Vorbereitung

### 1. GitHub-Benutzernamen/Organisationsnamen festlegen

Ersetzen Sie `your-org` durch Ihren tatsächlichen GitHub-Benutzernamen oder Organisationsnamen.

**Option A: Automatisch (empfohlen):**
```bash
# Auf Linux/Mac oder WSL
chmod +x scripts/prepare_github.sh
./scripts/prepare_github.sh YOUR_GITHUB_USERNAME
```

**Option B: Manuell**
Ersetzen Sie in folgenden Dateien `your-org` durch Ihren GitHub-Benutzernamen:
- `README.md` (3 Stellen)
- `installer/install_watcher.sh` (3 Stellen)
- `installer/watcher.service` (1 Stelle)
- `installer/watcher.service.example` (1 Stelle)

### 2. Repository auf GitHub erstellen

1. Gehen Sie zu https://github.com/new
2. Repository-Name: `prox-watch-monitoring-system`
3. Beschreibung: "Proxmox Monitoring System - Offline-First, Privacy-Preserving, Deterministic"
4. **Wichtig:** Repository als **privat** oder **öffentlich** wählen (je nach Bedarf)
5. **NICHT** "Initialize with README" aktivieren (Repository ist bereits initialisiert)
6. Klicken Sie auf "Create repository"

### 3. Lokales Repository vorbereiten

```bash
# Prüfen Sie, ob alle Dateien committed sind
git status

# Falls nicht, alle Dateien hinzufügen
git add .

# Ersten Commit erstellen (falls noch keine Commits vorhanden)
git commit -m "Initial commit: prox-watch monitoring system v0.5.0"

# Branch auf 'main' umbenennen (falls nötig)
git branch -M main
```

### 4. GitHub Remote hinzufügen

```bash
# Ersetzen Sie YOUR_GITHUB_USERNAME durch Ihren tatsächlichen Benutzernamen
git remote add origin https://github.com/YOUR_GITHUB_USERNAME/prox-watch-monitoring-system.git

# Remote prüfen
git remote -v
```

### 5. Zu GitHub pushen

```bash
# Ersten Push durchführen
git push -u origin main

# Bei privaten Repositories: GitHub wird nach Authentifizierung fragen
# Optionen:
# - Personal Access Token (empfohlen)
# - SSH-Key
# - GitHub CLI (gh auth login)
```

## Nach dem Upload

### Installationsbefehl testen

Auf dem Raspberry Pi:

```bash
# Standard-Installation (ohne Hardware-GPIO)
curl -fsSL https://raw.githubusercontent.com/YOUR_GITHUB_USERNAME/prox-watch-monitoring-system/main/installer/install_watcher.sh | sudo bash

# Mit Hardware-GPIO-Unterstützung
ENABLE_GPIO=1 curl -fsSL https://raw.githubusercontent.com/YOUR_GITHUB_USERNAME/prox-watch-monitoring-system/main/installer/install_watcher.sh | sudo bash
```

**Wichtig:** Ersetzen Sie `YOUR_GITHUB_USERNAME` durch Ihren tatsächlichen GitHub-Benutzernamen!

## Troubleshooting

### "Repository not found" (404)

- Prüfen Sie, ob das Repository auf GitHub existiert
- Prüfen Sie, ob Sie die richtigen Zugangsdaten verwenden
- Bei privaten Repositories: Prüfen Sie Ihre Berechtigungen

### "Permission denied"

- Verwenden Sie ein Personal Access Token statt Passwort
- Oder konfigurieren Sie SSH-Keys für GitHub

### Authentifizierung

**Personal Access Token erstellen:**
1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. "Generate new token (classic)"
3. Scopes: `repo` (für private Repos) oder `public_repo` (für öffentliche Repos)
4. Token kopieren und sicher aufbewahren
5. Bei `git push` als Passwort verwenden

## Nächste Schritte

Nach erfolgreichem Upload:
1. ✅ README.md aktualisieren (GitHub-URLs sollten bereits ersetzt sein)
2. ✅ Installationsbefehl auf Raspberry Pi testen
3. ✅ Repository als Template markieren (optional)
4. ✅ Releases/Tags erstellen (optional)
