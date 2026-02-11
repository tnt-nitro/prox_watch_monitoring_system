# Installer

Installation scripts and systemd service unit for prox-watch.

## Files

- `prox-watch.service` - systemd service unit definition

## Installation

### Manual Installation

1. Copy the service unit:
   ```bash
   sudo cp prox-watch.service /etc/systemd/system/
   ```

2. Reload systemd:
   ```bash
   sudo systemctl daemon-reload
   ```

3. Enable the service:
   ```bash
   sudo systemctl enable prox-watch.service
   ```

4. Start the service:
   ```bash
   sudo systemctl start prox-watch.service
   ```

### Prerequisites

Before installing the service, ensure:

1. User and group `prox-watch` exist:
   ```bash
   sudo useradd -r -s /usr/sbin/nologin prox-watch
   sudo groupadd -r prox-watch
   ```

2. Working directory exists:
   ```bash
   sudo mkdir -p /var/lib/prox-watch
   sudo chown prox-watch:prox-watch /var/lib/prox-watch
   sudo chmod 700 /var/lib/prox-watch
   ```

3. Binary is installed:
   ```bash
   sudo cp prox-watch /usr/local/bin/
   sudo chmod 755 /usr/local/bin/prox-watch
   ```

4. Configuration is initialized:
   ```bash
   sudo prox-watch init
   ```

## Service Management

### Status
```bash
sudo systemctl status prox-watch.service
```

### Logs
```bash
sudo journalctl -u prox-watch.service -f
```

### Restart
```bash
sudo systemctl restart prox-watch.service
```

### Stop
```bash
sudo systemctl stop prox-watch.service
```

## Security

The service unit includes security hardening:

- `NoNewPrivileges=true` - Prevents privilege escalation
- `PrivateTmp=true` - Isolated /tmp
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=true` - No access to home directories
- `ReadWritePaths=/var/lib/prox-watch` - Minimal write access
- `ReadOnlyPaths=/var/log/journal` - Journal read access only
- `CapabilityBoundingSet=` - No capabilities
- `RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6` - Limited network access
- `RestrictNamespaces=true` - No namespace creation
- Process limits (NOFILE=1024, NPROC=64)

## Configuration

The service runs as user `prox-watch` with working directory `/var/lib/prox-watch`.

Configuration file: `/var/lib/prox-watch/config.yaml`
State database: `/var/lib/prox-watch/state.db`
Secrets file: `/var/lib/prox-watch/secrets.yaml`

## Troubleshooting

### Service fails to start

1. Check logs:
   ```bash
   sudo journalctl -u prox-watch.service -n 50
   ```

2. Verify configuration:
   ```bash
   sudo prox-watch status
   ```

3. Check permissions:
   ```bash
   ls -la /var/lib/prox-watch
   ```

### Permission denied

Ensure the `prox-watch` user owns the working directory:
```bash
sudo chown -R prox-watch:prox-watch /var/lib/prox-watch
```
