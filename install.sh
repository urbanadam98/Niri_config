#!/usr/bin/env bash
# Installs Niri + DMS setup from this dotfiles repo.
# Run from the repo root: bash install.sh

set -euo pipefail

REPO="$(cd "$(dirname "$0")" && pwd)"

ln() { command ln -sfn "$1" "$2"; }
info() { echo "  $*"; }

echo "==> Niri config"
mkdir -p ~/.config/niri/dms
ln "$REPO/niri/config.kdl"        ~/.config/niri/config.kdl
for f in "$REPO"/niri/dms/*.kdl; do
    ln "$f" ~/.config/niri/dms/"$(basename "$f")"
done

echo "==> DankMaterialShell"
mkdir -p ~/.config/DankMaterialShell
ln "$REPO/dms/settings.json"        ~/.config/DankMaterialShell/settings.json
ln "$REPO/dms/plugin_settings.json" ~/.config/DankMaterialShell/plugin_settings.json

echo "==> Kitty"
mkdir -p ~/.config/kitty
ln "$REPO/kitty/kitty.conf" ~/.config/kitty/kitty.conf

echo "==> Fuzzel"
mkdir -p ~/.config/fuzzel
ln "$REPO/fuzzel/fuzzel.ini" ~/.config/fuzzel/fuzzel.ini

echo "==> Starship"
ln "$REPO/starship.toml" ~/.config/starship.toml

echo "==> Vicinae"
mkdir -p ~/.config/vicinae
ln "$REPO/vicinae/settings.json" ~/.config/vicinae/settings.json

echo "==> Environment"
mkdir -p ~/.config/environment.d
ln "$REPO/environment.d/qt-theme.conf" ~/.config/environment.d/qt-theme.conf

echo "==> Scripts"
mkdir -p ~/.local/bin
install -m755 "$REPO/bin/tasknotes-dms-bridge" ~/.local/bin/tasknotes-dms-bridge
install -m755 "$REPO/bin/pomodoro-stats"        ~/.local/bin/pomodoro-stats
install -m755 "$REPO/bin/pomodoro-sync"         ~/.local/bin/pomodoro-sync

echo "==> Systemd"
mkdir -p ~/.config/systemd/user
ln "$REPO/systemd/tasknotes-dms-bridge.service" ~/.config/systemd/user/tasknotes-dms-bridge.service
systemctl --user daemon-reload
systemctl --user enable --now tasknotes-dms-bridge
info "tasknotes-dms-bridge enabled and started"

echo ""
echo "Done. Reload Niri config: niri msg action reload-config"
echo "Note: DMS, Niri, Obsidian (TaskNotes) must be installed separately."
