#!/usr/bin/env bash
set -euo pipefail

echo "Starting demo experiment..."
sleep 5

mkdir -p /home/users/ntu/phoe0012/sync/phoebe/demo_exp

cat > /home/users/ntu/phoe0012/sync/phoebe/demo_exp/run.py <<'PYEOF'
print("Post-processing demo...")
with open("output.txt", "w") as f:
    f.write("Demo output generated automatically.\n")
PYEOF

cat > /home/users/ntu/phoe0012/sync/phoebe/demo_exp/result.txt <<'EOF'
Accuracy: 92%
EOF

touch /home/users/ntu/phoe0012/sync/phoebe/demo_exp/stop.txt

echo "Demo experiment finished."
echo "result.txt and stop.txt created."
