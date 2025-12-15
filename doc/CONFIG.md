# Configuration Guide (`config.yaml`)

The watcher is fully controlled by a YAML configuration file.

- Docker uses: `config/config.docker.yaml`
- Local runs use: `config/config.local.yaml`

No code changes are required for normal usage.

---

## 1. Top-Level Structure

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
      quiescent_seconds: 10
```

---

## 2. Key Fields Explained

### `root_input_dir`

The root directory that the watcher monitors.

Inside Docker this is always:

```text
/sync
```

Mapped from the CPU server, for example:

```text
~/phase1_sync
```

---

### `poll_seconds`

How often the watcher scans for new experiments.

- Lower value → more responsive
- Higher value → lower CPU usage

---

### `users`

Each user entry defines a subfolder under `/sync`.

Example:

```yaml
- id: "phoebe"
  input_subdir: "phoebe"
```

This monitors:

```text
/sync/phoebe
```

---

### `experiment_pattern`

Controls which folders are treated as experiments.

| Pattern | Matches |
|--------|--------|
| `test_exp*` | test_exp01, test_exp_demo |
| `exp*` | expA, experiment2 |
| `*` | all folders |

---

### `stop_file`

Trigger file that signals an experiment is ready.

Default:

```text
stop.txt
```

---

### `runner`

Determines what the watcher executes.

| Mode | Behaviour |
|------|----------|
| `auto` | Runs notebook if present, else script |
| `notebook` | Always runs `analysis.ipynb` |
| `script` | Always runs `run.py` |

---

### `quiescent_seconds`

How long a folder must remain unchanged before execution.

Prevents execution while Syncthing is still syncing.

---

### Output configuration (optional)

```yaml
output_ipynb: "analysis_output.ipynb"
output_html: "analysis_report.html"
output_pdf: "analysis_report.pdf"
```

PDF generation requires additional system dependencies.

---

## 3. Full Example

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
      quiescent_seconds: 10
      output_ipynb: "analysis_output.ipynb"
      output_html: "analysis_report.html"
      output_pdf: "analysis_report.pdf"
```

---

## 4. Notes

- Users never modify watcher source code
- Adding or removing users is config-only
- Experiment naming is controlled via patterns
