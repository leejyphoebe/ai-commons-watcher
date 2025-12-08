#!/usr/bin/env python3
"""
AI-Commons experiment watcher (Option C: experiment.yaml optional).

- Watches a root_input_dir for user experiment folders.
- For each experiment:
  - waits until it is "quiet" (no recent writes),
  - requires a stop.txt file to signal "ready",
  - decides how to run it based on:
      1) experiment.yaml (if present), OR
      2) auto-detection (analysis.ipynb / run.py).
  - after running, removes stop.txt so it won't re-run.

Config file (YAML) layout:

report_watchers:
  root_input_dir: "/sync"
  poll_seconds: 10
  users:
    - id: "phoebe"
      input_subdir: "phoebe"
      experiment_pattern: "test_exp*"
      stop_file: "stop.txt"
      runner: "auto"            # default runner if no experiment.yaml
      quiescent_seconds: 10
      default_notebook: "analysis.ipynb"
      default_script: "run.py"
      output_ipynb: "analysis_output.ipynb"
      output_html: "analysis_report.html"
      output_pdf: "analysis_report.pdf"
"""

import argparse
import subprocess
import time
from pathlib import Path
from typing import Optional, Tuple

import yaml


# ----------------- Helpers: YAML & quiet-folder check ----------------- #

def load_yaml(path: Path) -> dict:
    if not path.exists():
        return {}
    with path.open("r") as f:
        return yaml.safe_load(f) or {}


def folder_quiet_for(exp_dir: Path, seconds: int) -> bool:
    """True if nothing in exp_dir changed in the last `seconds`."""
    now = time.time()
    latest_mtime = 0.0

    for p in exp_dir.rglob("*"):
        try:
            mtime = p.stat().st_mtime
        except FileNotFoundError:
            # File might disappear during scan, ignore
            continue
        if mtime > latest_mtime:
            latest_mtime = mtime

    # If folder is empty, treat as quiet
    if latest_mtime == 0.0:
        return True

    return (now - latest_mtime) >= seconds


# ----------------- Runner selection (Option C) ----------------- #

def decide_runner(
    exp_dir: Path,
    user_cfg: dict,
) -> Tuple[Optional[str], Optional[Path]]:
    """
    Decide how to run this experiment.

    Priority:
    1) experiment.yaml (if present)
    2) fallback to user defaults + auto-detect.

    Returns (runner_type, path_to_target) where runner_type in {"notebook","script"}.
    If cannot decide, returns (None, None).
    """
    exp_yaml = exp_dir / "experiment.yaml"
    exp_cfg = load_yaml(exp_yaml) if exp_yaml.exists() else {}

    # From experiment.yaml if present
    runner = exp_cfg.get("runner")
    notebook_name = exp_cfg.get("notebook")
    script_name = exp_cfg.get("script")

    # Fall back to user defaults
    if runner is None:
        runner = user_cfg.get("runner", "auto")

    if notebook_name is None:
        notebook_name = user_cfg.get("default_notebook", "analysis.ipynb")

    if script_name is None:
        script_name = user_cfg.get("default_script", "run.py")

    notebook_path = exp_dir / notebook_name
    script_path = exp_dir / script_name

    # Auto mode: prefer notebook if it exists, else script
    if runner == "auto":
        if notebook_path.exists():
            return "notebook", notebook_path
        elif script_path.exists():
            return "script", script_path
        else:
            print(f"{exp_dir}: auto runner but no {notebook_name} or {script_name} found.")
            return None, None

    # Explicit notebook
    if runner == "notebook":
        if not notebook_path.exists():
            print(f"{exp_dir}: runner=notebook but {notebook_name} not found.")
            return None, None
        return "notebook", notebook_path

    # Explicit script
    if runner == "script":
        if not script_path.exists():
            print(f"{exp_dir}: runner=script but {script_name} not found.")
            return None, None
        return "script", script_path

    print(f"{exp_dir}: unknown runner '{runner}'. Expected auto|notebook|script.")
    return None, None


# ----------------- Runners: notebook & script ----------------- #

def run_notebook(
    notebook_path: Path,
    exp_dir: Path,
    user_cfg: dict,
):
    """
    Execute a Jupyter notebook and export HTML + PDF.

    - Uses jupyter nbconvert.
    - Writes outputs inside exp_dir.
    - Fails gracefully if PDF export doesn't work (e.g., LaTeX missing).
    """
    output_ipynb = exp_dir / user_cfg.get("output_ipynb", "analysis_output.ipynb")
    output_html = exp_dir / user_cfg.get("output_html", "analysis_report.html")
    output_pdf = exp_dir / user_cfg.get("output_pdf", "analysis_report.pdf")

    # 1) Execute notebook -> analysis_output.ipynb
    print(f"[Notebook] Executing {notebook_path.name} in {exp_dir} ...")
    subprocess.run(
        [
            "jupyter", "nbconvert",
            "--to", "notebook",
            "--execute", str(notebook_path),
            "--output", str(output_ipynb),
        ],
        check=True,
        cwd=str(exp_dir),
    )

    # 2) Export HTML
    print(f"[Notebook] Exporting HTML -> {output_html.name}")
    subprocess.run(
        [
            "jupyter", "nbconvert",
            "--to", "html",
            str(output_ipynb),
            "--output", str(output_html),
        ],
        check=True,
        cwd=str(exp_dir),
    )

    # 3) Export PDF (best-effort; don't crash whole watcher)
    try:
        print(f"[Notebook] Exporting PDF -> {output_pdf.name}")
        subprocess.run(
            [
                "jupyter", "nbconvert",
                "--to", "pdf",
                str(output_ipynb),
                "--output", str(output_pdf),
            ],
            check=True,
            cwd=str(exp_dir),
        )
    except subprocess.CalledProcessError as e:
        print(f"PDF export failed for {exp_dir}: {e}. "
              f"HTML report is still available.")


def run_script(
    script_path: Path,
    exp_dir: Path,
):
    """
    Run a Python script in the experiment folder.

    The script is responsible for generating its own outputs.
    """
    print(f"[Script] Running {script_path.name} in {exp_dir} ...")
    subprocess.run(
        ["python", str(script_path)],
        check=True,
        cwd=str(exp_dir),
    )


# ----------------- Main loop ----------------- #

def load_config(config_path: Path) -> dict:
    cfg = load_yaml(config_path)
    rw = cfg.get("report_watchers")
    if not isinstance(rw, dict):
        raise ValueError("config.yaml must have 'report_watchers' as a mapping")

    root_input_dir = Path(rw["root_input_dir"]).expanduser()
    poll_seconds = int(rw.get("poll_seconds", 10))

    users = rw.get("users", [])
    if not users:
        raise ValueError("report_watchers.users must contain at least one user entry")

    # Normalize user paths and defaults
    for u in users:
        if "id" not in u:
            raise ValueError("Each user must have an 'id'")
        if "input_subdir" not in u:
            raise ValueError(f"user {u.get('id')} must have input_subdir")

        u.setdefault("experiment_pattern", "*")
        u.setdefault("stop_file", "stop.txt")
        u.setdefault("runner", "auto")
        u.setdefault("quiescent_seconds", 10)
        u.setdefault("default_notebook", "analysis.ipynb")
        u.setdefault("default_script", "run.py")
        u.setdefault("output_ipynb", "analysis_output.ipynb")
        u.setdefault("output_html", "analysis_report.html")
        u.setdefault("output_pdf", "analysis_report.pdf")

        # Pre-compute the base input_dir
        u["input_dir"] = root_input_dir / u["input_subdir"]

    return {
        "root_input_dir": root_input_dir,
        "poll_seconds": poll_seconds,
        "users": users,
    }


def process_experiment(exp_dir: Path, user_cfg: dict):
    """Handle a single experiment directory."""
    user_id = user_cfg["id"]
    stop_file = exp_dir / user_cfg["stop_file"]

    if not stop_file.exists():
        return  # not ready

    if not folder_quiet_for(exp_dir, int(user_cfg["quiescent_seconds"])):
        return  # still syncing

    print(f"\n Found experiment: {exp_dir} (user={user_id})")

    runner_type, target = decide_runner(exp_dir, user_cfg)
    if runner_type is None:
        print(f"Skipping {exp_dir}: unable to decide runner.")
        # Optional: remove stop to avoid re-trigger
        try:
            stop_file.unlink()
            print(f"Removed stop.txt from {exp_dir} after failure.")
        except Exception as e:
            print(f"Failed to remove stop.txt: {e}")
        return

    try:
        if runner_type == "notebook":
            run_notebook(target, exp_dir, user_cfg)
        elif runner_type == "script":
            run_script(target, exp_dir)
        else:
            print(f"Unknown runner type {runner_type} for {exp_dir}")
    except subprocess.CalledProcessError as e:
        print(f"Error processing {exp_dir}: {e}")
    finally:
        # Always remove stop.txt so we don't re-run endlessly
        if stop_file.exists():
            try:
                stop_file.unlink()
                print(f"Removed stop.txt from {exp_dir}")
            except Exception as e:
                print(f"Failed to remove stop.txt: {e}")


def main():
    parser = argparse.ArgumentParser(description="AI-Commons experiment watcher")
    parser.add_argument("--config", required=True, help="Path to config.yaml")
    args = parser.parse_args()

    cfg = load_config(Path(args.config))
    root_input_dir = cfg["root_input_dir"]
    poll_seconds = cfg["poll_seconds"]
    users = cfg["users"]

    print(f"Watcher started. Monitoring root_input_dir: {root_input_dir}")

    while True:
        for u in users:
            input_dir: Path = u["input_dir"]
            pattern: str = u["experiment_pattern"]

            if not input_dir.exists():
                # Avoid noisy spam
                continue

            for exp in input_dir.glob(pattern):
                if not exp.is_dir():
                    continue
                process_experiment(exp, u)

        time.sleep(poll_seconds)


if __name__ == "__main__":
    main()
EOF
