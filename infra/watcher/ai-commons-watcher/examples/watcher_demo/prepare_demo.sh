#!/usr/bin/env bash
set -euo pipefail

DEMO_DIR="/home/users/ntu/phoe0012/sync/phoebe/demo_exp"

mkdir -p "$DEMO_DIR"

rm -f "$DEMO_DIR/stop.txt" \
      "$DEMO_DIR"/stop.sync-conflict-*.txt

echo "Demo folder prepared: $DEMO_DIR"
