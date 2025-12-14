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

---

## 3. Repository Structure

```text
ai-commons-watcher/
│
├── Dockerfile
├── requirements.txt
├── README.md
│
├── watcher/
│   ├── __init__.py
│   └── watcher.py          # Core watcher logic
│
├── config/
│   ├── config.docker.yaml  # Config used inside Docker
│   └── config.local.yaml   # Config for local (non-Docker) testing
│
├── scripts/
│   └── install_syncthing_ubuntu.sh  # Helper script to install Syncthing
│
├── examples/
│   └── sample_experiment/
│       ├── experiment.yaml # Example experiment metadata
│       ├── run.py          # Example experiment script
│       └── stop.txt        # Trigger file
