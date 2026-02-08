# Developer Notes

This document explains the internal architecture and design decisions.

---

## 1. Architecture Overview

```text
Syncthing
   ↓
CPU Server
   ↓
Docker Watcher
   ↓
Run Experiment
   ↓
Generate Output
```

---

## 2. Core Watcher Algorithm (Simplified)

```text
while True:
  for each user:
    scan user directory
    for each experiment:
      if stop file exists:
        wait until folder is stable
        run experiment
        remove stop file
  sleep(poll_seconds)
```

---

## 3. Code Structure

### `watcher/watcher.py`

Responsibilities:
- YAML loading
- Folder scanning
- Quiet-folder detection
- Script execution
- Notebook execution
- Output generation
- Stop-file removal
- Error handling

---

## 4. Extending the System

### Adding a new runner

```python
def run_custom(experiment_path, cfg):
    ...
```

Dispatch logic:

```python
if runner == "custom":
    run_custom(...)
```

---

### Adding post-processing steps

```python
generate_summary(exp_dir)
upload_results(exp_dir)
```

---

### Adding new output formats

```python
generate_markdown()
generate_json_summary()
```

---

## 5. Known Constraints

- PDF generation requires system dependencies
- Notebook kernels must exist inside Docker
- Large files may require longer quiescent times

---

## 6. Testing Strategy

### Current
- Local watcher test
- Docker watcher test
- Multi-experiment test

### Future
- Unit tests for scanning
- Trigger detection
- Runner dispatch

---

## 7. Design Rationale

- Event-driven → efficient and predictable
- Configurable → user-friendly
- Dockerised → reproducible
- AI-Commons independent → standalone usable
