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

## 2. How It Works (Short Overview)

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

## 3. Minimal Local Test (Without Docker)

```bash
pip install -r requirements.txt
python -m watcher.watcher --config config/config.local.yaml
```

The watcher will start and wait for experiments.

---

## 4. Running in Docker (Recommended)

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

## 5. Creating an Example Experiment

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

## 6. Configuration Summary (Short)

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

- This watcher runs on the NTU CPU server
- No NSCC configuration is needed
- Syncthing setup is performed separately by each user
- Designed to integrate smoothly with the AI-Commons workflow



# AI-Commons Watcher

This repo provides a **Watcher** for AI-Commons experiments.

It:
- Watches a Syncthing sync folder (mounted as `/sync` in Docker).
- Waits for experiment folders + `stop.txt`.
- Runs either a **notebook** or a **Python script**.
- Produces HTML (+ optionally PDF) reports.
- Removes `stop.txt` after running so the experiment is not re-run.

---

## 1. Folder layout on CPU server

Example on the NTU CPU server:

/home/USER/phase1_sync/
  phoebe/
    test_exp_01/
    test_exp_02/
  silvi/
  ...

When running in Docker, this is mounted as:

/sync/
  phoebe/
  silvi/
  ...

The watcher only cares about what appears under `/sync/USERNAME`.

---

## 2. Config

Config file: `config/config.docker.yaml` (inside container)

```yaml
report_watchers:
  root_input_dir: "/sync"
  poll_seconds: 10

  users:
    - id: "phoebe"
      input_subdir: "phoebe"          # watches /sync/phoebe
      experiment_pattern: "test_exp*"
      stop_file: "stop.txt"
      runner: "auto"
      quiescent_seconds: 10

      default_notebook: "analysis.ipynb"
      default_script: "run.py"

      output_ipynb: "analysis_output.ipynb"
      output_html: "analysis_report.html"
      output_pdf: "analysis_report.pdf"

```

---

## 3. Repository Structure

```text
ai-commons-watcher/
  Dockerfile
  requirements.txt
  README.md

  watcher/
    __init__.py
    watcher.py          # Core watcher logic

  config/
    config.docker.yaml  # Config used inside Docker
    config.local.yaml   # Config for local (non-Docker) testing

  scripts/
    install_syncthing_ubuntu.sh  # Helper script to install Syncthing

  examples/
    sample_experiment/
      experiment.yaml   # Example experiment metadata
      run.py            # Example experiment script
      stop.txt          # Trigger file
```

---

## 4. How the Watcher Works (Execution Flow)

Syncthing synchronises experiment folders into the CPU server.

The watcher continuously monitors `/sync/USERNAME`.

When an experiment folder contains `stop.txt`, it is treated as ready.

The watcher waits until the folder is quiescent (no file changes).

The watcher executes:
- a notebook (`analysis.ipynb`), or
- a Python script (`run.py`),
depending on configuration.

Output files (HTML and optional PDF) are generated in the same folder.

`stop.txt` is removed to prevent re-processing.

---

## 5. Local Testing (Without Docker)

This verifies watcher logic before containerisation.

```bash
pip install -r requirements.txt
python -m watcher.watcher --config config/config.local.yaml
```

Expected output:
- Watcher starts successfully
- Displays monitored directory
- Waits for experiments

---

## 6. Docker-Based Testing (Recommended)

Build the Docker image:

```bash
docker build -t ai-commons-watcher .
```

Run the watcher container:

```bash
docker run -v ~/phase1_sync:/sync ai-commons-watcher \
  --config /app/config/config.docker.yaml
```

Expected behaviour:
- `/home/USER/phase1_sync` is mounted as `/sync`
- Experiments with `stop.txt` are automatically executed
- Reports are generated
- Completed experiments are not re-run

---

## 7. Creating a Test Experiment

```bash
mkdir -p ~/phase1_sync/phoebe/test_exp_demo
cd ~/phase1_sync/phoebe/test_exp_demo
```

Create a script:

```bash
cat << 'EOF' > run.py
print("Hello from test experiment")
with open("result.txt", "w") as f:
    f.write("Experiment completed.")
EOF
```

Trigger execution:

```bash
touch stop.txt
```

The watcher should:
- Detect the experiment
- Execute `run.py`
- Produce output files
- Remove `stop.txt`

---

## 8. Design Notes

The watcher is client-side and runs on the NTU CPU server.

No NSCC configuration is required for watcher execution.

Syncthing setup is user-managed to allow flexible deployment.

Docker ensures a reproducible runtime environment.

The system is compatible with AI-Commons but does not depend on it.

---

## 9. Customising the Watcher (What Users Can Change)

This README already explains what the system is, how it works, and how to run it locally and in Docker.

What this section covers is **what users can customise**, without modifying any source code.

All customisation is done via the configuration file.

---

### 9.1 Changing the monitored directory

In `config/config.docker.yaml`:

```yaml
root_input_dir: "/sync"
```

This is the root folder being monitored inside Docker.

By default, `/sync` is mapped to `~/phase1_sync` on the CPU server.

To change where experiments are read from, update the Docker volume mount:

```bash
docker run -v /NEW/PATH:/sync ai-commons-watcher \
  --config /app/config/config.docker.yaml
```

---

### 9.2 Adding or removing users

Each user corresponds to a subfolder under `/sync`.

Example configuration:

```yaml
users:
  - id: "phoebe"
    input_subdir: "phoebe"
```

To add another user:

```yaml
users:
  - id: "phoebe"
    input_subdir: "phoebe"

  - id: "silvi"
    input_subdir: "silvi"
```

The watcher will then monitor:

```
/sync/phoebe
/sync/silvi
```

---

### 9.3 Changing experiment naming rules

```yaml
experiment_pattern: "test_exp*"
```

This controls which folders are treated as experiments.

Examples:

- `"test_exp*"` → `test_exp_01`, `test_exp_demo`
- `"exp*"` → `exp1`, `experimentA`
- `"*"` → any folder

---

### 9.4 Changing the trigger file

```yaml
stop_file: "stop.txt"
```

This file signals that an experiment is ready to run.

Users may rename it if needed:

```yaml
stop_file: "READY"
```

Then trigger execution with:

```bash
touch READY
```

---

### 9.5 Choosing what gets executed

```yaml
runner: "auto"
default_notebook: "analysis.ipynb"
default_script: "run.py"
```

Supported modes:

- `auto` → runs notebook if present, otherwise script
- `notebook` → always runs `analysis.ipynb`
- `script` → always runs `run.py`

Example (script-only workflow):

```yaml
runner: "script"
```

---

### 9.6 Adjusting stability and polling behaviour

```yaml
poll_seconds: 10
quiescent_seconds: 10
```

- `poll_seconds`: how often the watcher scans folders
- `quiescent_seconds`: how long the folder must be unchanged before execution

Increase these values for:
- Large files
- Slow network synchronisation
- Cloud-based storage

---

### 9.7 Enabling or disabling PDF generation

```yaml
output_pdf: "analysis_report.pdf"
```

PDF generation is optional.

To disable PDF output, remove or comment out this line.

HTML output will still be generated.

---

## 10. What Users Do Not Need to Configure

- No changes are required on NSCC
- No AI-Commons installation is required to run the watcher
- No modification to watcher source code is needed for normal use

---

## 11. Typical User Workflow (Summary)

1. Sync experiment folder using Syncthing
2. Place `run.py` or `analysis.ipynb` inside the experiment folder
3. Create `stop.txt`
4. Watcher detects → executes → generates outputs
5. `stop.txt` is removed automatically

