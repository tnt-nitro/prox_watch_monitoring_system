#!/bin/bash
# Skript zum Ersetzen von "your-org" durch tatsächlichen GitHub-Benutzernamen
# 
# Verwendung:
#   ./scripts/prepare_github.sh YOUR_GITHUB_USERNAME

set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Fehler: GitHub-Benutzername fehlt"
    echo "Verwendung: $0 YOUR_GITHUB_USERNAME"
    exit 1
fi

GITHUB_USER="$1"

echo "Ersetze 'your-org' durch '$GITHUB_USER'..."

# Dateien mit Platzhaltern finden und ersetzen
find . -type f \( -name "*.md" -o -name "*.sh" -o -name "*.service" -o -name "*.example" \) \
    ! -path "./.git/*" \
    ! -path "./node_modules/*" \
    ! -path "./vendor/*" \
    -exec sed -i "s/your-org/$GITHUB_USER/g" {} +

echo "✅ Ersetzung abgeschlossen!"
echo ""
echo "Bitte prüfen Sie die Änderungen mit:"
echo "  git diff"
echo ""
echo "Dann committen und zu GitHub pushen:"
echo "  git add ."
echo "  git commit -m 'Update GitHub URLs'"
echo "  git push origin main"
