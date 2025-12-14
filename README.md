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
