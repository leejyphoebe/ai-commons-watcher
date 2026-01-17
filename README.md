# AI-Commons Watcher

A lightweight, event-driven watcher that automatically executes user experiments synced into the NTU CPU server via **Dockerised Syncthing**.

The watcher:

- Monitors a Syncthing folder (mounted as `/sync` in Docker)
- Detects experiment folders containing `stop.txt`
- Executes either a Python script (`run.py`) or notebook (`analysis.ipynb`)
- Generates HTML (and optional PDF) reports
- Removes `stop.txt` to prevent re-processing

---

## 1. Folder layout (Host Server)

```text
/path/to/ai-commons-watcher/
  sync/
    <your_username>/
      <any_experiment_folder_name>/
        run.py OR analysis.ipynb
        stop.txt

Inside Docker, `./sync` is mounted as `/sync`, so the watcher reads:

/sync/<your_username>/<any_experiment_folder_name>/
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

## 3. Prerequisites

Before running the watcher, you must:

- Access to an NTU CPU server
- Docker and Docker Compose available on the CPU server
- Git

> Syncthing is **not installed on the host OS** for the demo.
> It runs inside Docker using Docker Compose.

---

## 4. Docker-First Setup (Syncthing + Watcher)

This demo runs **both Syncthing and the watcher inside Docker** using Docker Compose.

### 4.1 Clone the repository

```bash
git clone <REPO_URL>
cd ai-commons-watcher
```

### 4.2 Configure permissions (important)

```bash
cp .env.example .env
# edit .env:
# PUID=$(id -u)
# PGID=$(id -g)
```
This ensures files created by Docker are owned by your user.


### 4.3 Configure watcher

```bash
cp config/config.template.yaml config/config.docker.yaml
# edit id and input_subdir to your NTU username
```

### 4.4 Start Syncthing + Watcher

```bash
mkdir -p sync
docker compose up -d --build
docker logs -f ai-commons-watcher
```

### 4.5 Syncthing Web UI and Ports

On shared NTU servers, default Syncthing ports may already be in use.
This project maps Syncthing to alternative host ports:

- Web UI: http://<host-server-ip>:28384
- Sync port: tcp://<host-server-ip>:22200

If inbound ports are blocked, use SSH port forwarding:

```bash
ssh -L 28384:localhost:28384 <user>@<host-server>
```
Then open http://localhost:28384 in your browser.

---

## 5. Creating an Example Experiment

All experiments must be created inside the `sync/` directory of this repository.

### 5.1 Create an experiment folder

```bash
mkdir -p sync/<your_username>/example_experiment
cd sync/<your_username>/example_experiment
```
You may name the experiment folder anything.

### 5.2 Script-based experiment (run.py)

```bash
cat << 'EOF' > run.py
print("Hello from experiment")
with open("result.txt", "w") as f:
    f.write("Success")
EOF
```
Trigger execution by creating the stop file:
```bash
touch stop.txt
```

### 5.3 Expected outcome

After detection, the watcher will:
- Execute `run.py`
- Generate output files (e.g. `result.txt`)
- Remove `stop.txt` to prevent re-running

The experiment folder will now contain the results.

---

## 6. Configuration Summary

Key settings from `config/config.docker.yaml`:

```yaml
report_watchers:
  root_input_dir: "/sync"
  poll_seconds: 10

users:
  - id: "YOUR_NTU_USERNAME"
    input_subdir: "YOUR_NTU_USERNAME"
    experiment_pattern: "*"
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

## 7. Troubleshooting

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

## 8. Notes

- This project is designed for NTU CPU servers
- Syncthing and the watcher run fully inside Docker
- No OS-level Syncthing installation is required for the demo
- Users may create experiment folders with any name
- Trigger execution by adding `stop.txt`
