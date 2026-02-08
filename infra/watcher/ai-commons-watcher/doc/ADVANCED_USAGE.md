# Advanced Usage Guide

This guide covers multi-user setups, power-user options, and non-default workflows.

---

## 1. Multi-User Monitoring

```yaml
users:
  - id: "phoebe"
    input_subdir: "phoebe"

  - id: "silvi"
    input_subdir: "silvi"
```

Watcher monitors:

```text
/sync/phoebe
/sync/silvi
```

Each user’s experiments are processed independently.

---

## 2. Notebook-Only or Script-Only Workflows

Script-only:

```yaml
runner: "script"
```

Notebook-only:

```yaml
runner: "notebook"
```

Automatic detection (recommended):

```yaml
runner: "auto"
```

---

## 3. Custom Trigger Filenames

Change trigger name:

```yaml
stop_file: "READY"
```

Trigger execution with:

```bash
touch READY
```

---

## 4. Handling Large Experiments

Increase timing thresholds:

```yaml
poll_seconds: 20
quiescent_seconds: 30
```

Useful for:
- Large files
- Slow network synchronisation
- Cloud storage

---

## 5. Changing the Syncthing Directory

```bash
docker run -v /data/sync_lab:/sync ai-commons-watcher \
  --config /app/config/config.docker.yaml
```

Ensure config still contains:

```yaml
root_input_dir: "/sync"
```

---

## 6. Testing Without Syncthing

```bash
mkdir -p ~/phase1_sync/user/expABC
echo 'print("hi")' > ~/phase1_sync/user/expABC/run.py
touch ~/phase1_sync/user/expABC/stop.txt
```

Watcher behaviour is identical.

---

## 7. When Nothing Happens

Reset trigger:

```bash
rm stop.txt
sleep 1
touch stop.txt
```

---

## 8. Debugging Docker Mounts

```bash
docker exec -it <container> ls -R /sync
```

- If files appear → watcher is working
- If not → fix volume mount
