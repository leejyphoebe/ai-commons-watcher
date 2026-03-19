#!/usr/bin/env bash
set -euo pipefail

SESSION_NAME="syncthing"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/syncthing"
GUI_ADDR="127.0.0.1:8384"
ST_VER="${ST_VER:-v2.0.13}"
SYNC_USER="${SYNC_USER:-$USER}"
SYNC_ROOT="$HOME/sync"
SYNC_DIR="$SYNC_ROOT/$SYNC_USER"

echo "==> AI-Commons Watcher: NSCC Syncthing bootstrap"

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)
    ST_ARCH="amd64"
    ;;
  aarch64|arm64)
    ST_ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

mkdir -p "$INSTALL_DIR"
mkdir -p "$SYNC_DIR"

if [ ! -x "$BINARY_PATH" ]; then
  echo "==> Installing Syncthing $ST_VER for linux-$ST_ARCH"
  TMP_DIR="$(mktemp -d)"
  TAR_NAME="syncthing-linux-${ST_ARCH}-${ST_VER}.tar.gz"
  URL="https://github.com/syncthing/syncthing/releases/download/${ST_VER}/${TAR_NAME}"

  curl -L "$URL" -o "$TMP_DIR/$TAR_NAME"
  tar -xzf "$TMP_DIR/$TAR_NAME" -C "$TMP_DIR"
  cp "$TMP_DIR"/syncthing-linux-${ST_ARCH}-${ST_VER}/syncthing "$BINARY_PATH"
  chmod +x "$BINARY_PATH"
  rm -rf "$TMP_DIR"
else
  echo "==> Syncthing already installed at $BINARY_PATH"
fi

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is not available on this system."
  echo "Please install or use an environment on NSCC where tmux is available."
  exit 1
fi

if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
  echo "==> tmux session '$SESSION_NAME' already exists"
else
  echo "==> Creating tmux session '$SESSION_NAME'"
  tmux new-session -d -s "$SESSION_NAME"
fi

if tmux capture-pane -pt "$SESSION_NAME" | grep -qi "syncthing"; then
  echo "==> Syncthing may already be running in tmux session '$SESSION_NAME'"
else
  echo "==> Starting Syncthing in tmux session '$SESSION_NAME'"
  tmux send-keys -t "$SESSION_NAME" "$BINARY_PATH --gui-address=$GUI_ADDR" C-m
fi

echo
echo "Setup complete."
echo
echo "Syncthing binary:"
echo "  $BINARY_PATH"
echo
echo "Shared experiment folder on NSCC:"
echo "  $SYNC_DIR"
echo
echo "IMPORTANT:"
echo "Create experiments inside:"
echo "  $SYNC_DIR/<experiment_name>"
echo
echo "Example:"
echo "  mkdir -p $SYNC_DIR/demo_exp_01"
echo
echo "Do NOT create experiments directly under:"
echo "  $SYNC_ROOT/<experiment_name>"
echo
echo "Next steps:"
echo "1. From your laptop, create an SSH tunnel:"
echo "   ssh -N -L 8384:127.0.0.1:8384 <nscc_username>@aspire2antu.nscc.sg"
echo
echo "2. Open in browser:"
echo "   http://127.0.0.1:8384"
echo
echo "3. In Syncthing on NSCC, copy the device ID from:"
echo "   Actions -> Show ID"
echo
echo "4. On the host Syncthing GUI, add the NSCC device"
echo
echo "5. Accept the shared folder on NSCC and keep the folder path as:"
echo "   $SYNC_DIR"
echo
echo "Useful commands:"
echo "  tmux attach -t $SESSION_NAME"
echo "  tmux kill-session -t $SESSION_NAME"
echo "  pkill -u \"$USER\" syncthing"
echo "  $BINARY_PATH --version"
echo
echo "To verify folder path:"
echo "  find $SYNC_ROOT -maxdepth 2 -type d | sort"
