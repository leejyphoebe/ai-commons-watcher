# AI-Commons Watcher

A lightweight, event-driven watcher that automatically executes user experiments synced into the NTU CPU server via **Dockerised Syncthing**.

The watcher:

- Monitors a Syncthing folder (mounted as `/sync` in Docker)
- Detects experiment folders containing `stop.txt`
- Executes either a Python script (`run.py`) or notebook (`analysis.ipynb`)
- Generates HTML (and optional PDF) reports
- Removes `stop.txt` to prevent re-processing

## Current Working Pipeline

The verified working flow is:
```
NSCC experiment folder
        ↓
Syncthing sync
        ↓
Host machine sync/<username>/<experiment_name>
        ↓
Watcher container detects stop.txt
        ↓
runs run.py or analysis.ipynb
        ↓
generates outputs
        ↓
sends email notification
        ↓
removes stop.txt
```

This pipeline has been tested end-to-end using NSCC, Syncthing, Docker, and the AI-Commons Watcher.


## Setup Overview

This project requires two one-time setups, which **must be completed in order**:

### 1. CPU Server (Host Server) — *Required First*
Runs Docker-based services (AI-Commons Watcher + Syncthing) and is intended to be always-on.

- Hosts the Syncthing service inside Docker
- Accepts inbound Syncthing connections
- Executes synced experiments automatically

**Important**  
Due to NSCC firewall restrictions, all Syncthing device connections must be initiated from the Host Server.

### 2. NSCC (User Side) — *After Host Setup*
Runs Syncthing as a user-level process to sync experiment folders to the CPU server.
This setup is performed **once per user**.

- NSCC does **not** accept inbound connections
- NSCC only accepts device and folder requests from the Host Server

**Important**  
Do not attempt to add the Host Server as a remote device from the NSCC Syncthing GUI.  
All Syncthing connections are initiated from the Host Server.

If you are syncing experiments from **NSCC**, you must complete the NSCC setup **once**.

**Follow this guide carefully:**  
[NSCC Syncthing Setup Guide](docs/nscc-syncthing-setup.md)

### Important: Correct NSCC Experiment Folder

Experiments must be created inside the accepted Syncthing shared folder:
```bash
~/sync/<username>/<experiment_name>
```
Example:
```bash
~/sync/phoe0012/demo_exp_01
```
Do **NOT** create experiments directly under:
```bash
~/sync/demo_exp_01
```

This guide covers:
- Installing Syncthing without sudo
- Running Syncthing persistently using `tmux`
- Accessing the Syncthing GUI (for initial setup only)
- Ensuring automatic background syncing to the CPU server

### One-Time Setup Philosophy

The intended workflow is:

1. User sets up Syncthing on NSCC once
2. Syncthing is launched inside `tmux`
3. User detaches and logs out
4. All future experiment files sync automatically
5. The watcher on the CPU server handles execution without user action

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

On the host machine, check your SYNC_DIR folder instead
(e.g. `~/ai-commons-sync`).

### Syncthing Volume Layout

Syncthing runs inside Docker and uses two separate mounts:
```bash
./sync → /var/syncthing/data
./syncthing-config → /var/syncthing/config
```

Purpose:

| Host Folder | Purpose |
|-------------|---------|
| `sync/` | Synced experiment folders |
| `syncthing-config/` | Syncthing device and folder configuration |

Separating data and configuration ensures that Syncthing settings are **not lost when Docker containers restart**.

---

## 2. How It Works

- Syncthing syncs experiment folders from NSCC into the CPU server
- Docker watcher monitors `/sync/<USERNAME>`
- When a folder contains `stop.txt`, the watcher:
  - waits until the folder stops changing
  - runs either:
    - `analysis.ipynb`, or
    - `run.py`
  - generates output files (HTML / PDF / text outputs)
  - creates `outputs.zip`
  - sends an email notification to `EMAIL_TO`
  - removes `stop.txt`
- The watcher continues running in the background

This allows users to receive experiment results on their phone without logging into the server.

## Syncthing Runtime Model (Important)

This project uses Syncthing to sync experiment folders from NSCC to the NTU CPU server.

### CPU Server (Host Server)
- Syncthing runs **inside Docker**
- Managed by Docker Compose
- Configured with `restart: unless-stopped`
- Intended to be **always running**
- No user interaction required after setup

### NSCC
- Syncthing runs as a **user-level process**
- System services (systemd) are not available on NSCC
- Syncthing must be started **once inside a tmux session**
- After setup, users **do not need to touch Syncthing again**
- File syncing happens automatically in the background

This design follows standard HPC usage patterns and NSCC constraints.


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
> **WSL users (Windows):**  
> Run all Docker commands inside the Ubuntu (WSL) terminal.
> Docker Desktop must be running on Windows.


### 4.1 Clone the repository

```bash
git clone <REPO_URL>
cd ai-commons-watcher
```

### 4.2 Configure `.env` (required)

Create your local environment file:

```bash
cp .env.example .env
nano .env
```

Update the following fields:
```env
PUID=1000
PGID=1000
EMAIL_TO=your_email@example.com
```

How to get PUID and PGID on Linux / WSL:
```bash
id -u
id -g
```
Use the returned values in .env.

#### What each field means

PUID = your Linux user ID

PGID = your Linux group ID

EMAIL_TO = the email address that should receive experiment notifications

### Important

This project uses a **shared sender email account** for experiment notifications.

Each user runs the watcher on their **own host server**, so the shared SMTP
credentials must be added to that server's local `.env` file.

You will need to obtain the password from the project maintainer:

- SMTP_PASS

Users only need to set their own recipient email address:

EMAIL_TO=your_email@example.com

Do not commit real SMTP credentials into Git.

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

## 5. Creating an Example Experiment (Host or Pre-Shared Folder)

In practice, experiment folders are typically created on the Host Server and shared to NSCC automatically via Syncthing.

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

- execute `run.py` or `analysis.ipynb`
- generate output files (for example `result.txt`, HTML, or PDF)
- create `outputs.zip`
- send a completion email to `EMAIL_TO`
- remove `stop.txt` to prevent re-running

The experiment folder will then contain the generated results, and the user will receive an email notification with attachments when available.

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

### Watcher shows no output?

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

### Watcher does not detect `stop.txt` even though files exist

If the watcher is running but does not react to `stop.txt`, the most common cause
is an incorrect Docker volume mount.

Verify what the watcher container is actually mounting:

```bash
docker inspect ai-commons-watcher \
  --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}'
```

Ensure that the host sync/ directory is mounted to /sync inside the container.
Then verify from inside the container:

```bash
docker exec -it ai-commons-watcher ls -R /sync
```

If your experiment folder does not appear under /sync/<your_username>/,
the watcher will not detect it.

---

## 8. Notes

- This project is designed for NTU CPU servers
- Syncthing and the watcher run fully inside Docker
- No OS-level Syncthing installation is required for the demo
- Users may create experiment folders with any name (within synced folders)
- Trigger execution by adding `stop.txt`
  

## Email Notifications

The watcher can automatically send an email when an experiment finishes.

This allows users to:
- know when post-processing is complete
- view the result on their phone
- download attached output files without logging into the server

### How email works

The watcher uses a shared sender account that is already configured for the deployment.

Users only need to set their own recipient email address in `.env`:

```bash
cp .env.example .env
nano .env
```
Then update:
```bash
EMAIL_TO=your_email@example.com
```
You do not need to configure your own SMTP account or app password.

### Email Notification Flow

```
Experiment finishes
        ↓
stop.txt detected
        ↓
Watcher runs script / notebook
        ↓
Outputs generated
        ↓
outputs.zip created
        ↓
Email notification sent to EMAIL_TO
        ↓
stop.txt removed
```

### What the email may contain

The watcher may attach:

the main output file, such as:
- analysis_report.pdf
- analysis_report.html
- result.txt
- report.txt
- error.log
- outputs.zip, which contains experiment outputs

### If no email is received

Check the watcher logs:
```
docker logs -f ai-commons-watcher
```
Common causes:
- EMAIL_TO was not set in .env
- the .env file was not created before running Docker Compose
- sender SMTP credentials were missing or invalid
