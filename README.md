# AI-Commons Watcher

A lightweight, event-driven watcher that automatically executes user experiments synced into the NTU CPU server via Syncthing.

The watcher:

- Monitors a Syncthing folder (mounted as `/sync` in Docker)
- Detects experiment folders containing `stop.txt`
- Executes either a Python script (`run.py`) or notebook (`analysis.ipynb`)
- Generates HTML (and optional PDF) reports
- Removes `stop.txt` to prevent re-processing

---

## 1. Repository Structure

```text
ai-commons-watcher/
  Dockerfile
  requirements.txt
  README.md

  watcher/
    watcher.py          # Main watcher logic

  config/
    config.docker.yaml  # Config used inside Docker
    config.local.yaml   # Config for local testing

  scripts/
    install_syncthing_ubuntu.sh  # Optional helper for setting up Syncthing

  examples/
    sample_experiment/
      run.py
      experiment.yaml
      stop.txt
```

---

## 2. How It Works

- Syncthing syncs user experiments into the CPU server
- Docker watcher monitors `/sync/<USERNAME>`
- When a folder contains `stop.txt`, the watcher:
  - waits until the folder stops changing
  - runs either:
    - `analysis.ipynb`, or
    - `run.py`
  - generates output files (HTML / PDF)
  - removes `stop.txt`
- The watcher continues running in the background

This design ensures automated, repeatable, hands-free experiment processing.

---

---

## 3. Prerequisites

Before running the watcher, you must:

- Have access to the NTU CPU server
- Have Docker installed on the CPU server
- Install and configure Syncthing
- Clone this repository onto the CPU server

### 3.1 Clone the Repository

On the NTU CPU server, clone this repository:

```bash
git clone <REPO_URL>
cd ai-commons-watcher
```

All commands in this README assume you are inside the repository root.

---

## 4. Syncthing Setup

> **Important**
>
> The watcher can start without Syncthing, but it will see an empty `/sync` directory.
>
> To actually process experiments, Syncthing **must be set up first** so that experiment
> folders are synced into the CPU server before running Docker.

This repository includes a helper script for installing Syncthing
on Ubuntu / Debian / WSL:

```text
scripts/install_syncthing_ubuntu.sh
```

### 4.1 Install Syncthing

On the NTU CPU server (and optionally on your laptop):

```bash
chmod +x scripts/install_syncthing_ubuntu.sh
./scripts/install_syncthing_ubuntu.sh
```

This script:
- Adds the official Syncthing APT repository
- Imports the GPG signing key
- Installs the latest stable Syncthing release

---

### 4.2 Start Syncthing (CLI Mode)

Start Syncthing on the **CPU server**:

```bash
syncthing serve --no-browser --no-restart &
```

Start Syncthing on your **laptop** (or NSCC machine):

```bash
syncthing serve --no-browser --no-restart &
```

Once running, Syncthing will generate a **device ID** for each machine.

---

### 4.3 How Syncthing Integrates With the Watcher

After setup, Syncthing ensures:

```text
Laptop / NSCC → CPU server → Docker
```

Folder mapping:

```text
CPU server folder: ~/phase1_sync/USERNAME
Docker view:       /sync/USERNAME
```

The watcher monitors `/sync/USERNAME` and executes experiments automatically
when `stop.txt` appears.


## 5. Minimal Local Test (Without Docker)

```bash
pip install -r requirements.txt
python -m watcher.watcher --config config/config.local.yaml
```

The watcher will start and wait for experiments.

---

## 6. Running in Docker (Recommended)

Build the image:

```bash
docker build -t ai-commons-watcher .
```

Run the watcher:

```bash
docker run \
  -v ~/phase1_sync:/sync \
  ai-commons-watcher \
  --config /app/config/config.docker.yaml
```

Where:
- `~/phase1_sync` is your Syncthing root
- Inside Docker it becomes `/sync`

---

## 7. Creating an Example Experiment

```bash
mkdir -p ~/phase1_sync/USER/test_exp_demo
cd ~/phase1_sync/USER/test_exp_demo
```

Add a script:

```bash
cat << 'EOF' > run.py
print("Hello from experiment")
with open("result.txt", "w") as f:
    f.write("Success")
EOF
```

Trigger execution:

```bash
touch stop.txt
```

Expected outcome:
- The watcher detects the folder
- Runs `run.py`
- Produces `result.txt`
- Removes `stop.txt`

---

## 8. Configuration Summary

Key settings from `config/config.docker.yaml`:

```yaml
report_watchers:
  root_input_dir: "/sync"
  poll_seconds: 10

  users:
    - id: "phoebe"
      input_subdir: "phoebe"
      experiment_pattern: "test_exp*"
      stop_file: "stop.txt"
      runner: "auto"
```

Users can easily adjust:
- which folders are monitored
- experiment naming patterns
- trigger filename
- script / notebook runner mode

Full customisation is documented separately.

---

## 9. Troubleshooting

Watcher shows no output?

Touch the trigger again:

```bash
rm -f stop.txt
sleep 1
touch stop.txt
```

Check that Docker sees your files:

```bash
docker exec -it <container> ls -R /sync
```

Check logs:

```bash
docker logs -f <container>
```

---

## 10. Notes

- This watcher runs on the NTU CPU server
- No NSCC configuration is needed
- Syncthing setup is performed separately by each user
- Designed to integrate smoothly with the AI-Commons workflow
