# NSCC Syncthing Setup Guide (User-Level, No Sudo)

This guide explains how to set up **Syncthing on NSCC** to sync experiment data to the host server.

## 1. Prerequisites

This guide assumes the **Host Server has already been set up**.

- NSCC account with SSH access
- Working internet connection on NSCC
- Host Server Syncthing running inside Docker

---

## 2. Check if Syncthing Is Already Installed

On NSCC, run:

```bash
which syncthing
```

---

## 3. Install Syncthing (User-Level, No Sudo)

### 3.1 Confirm Architecture

```bash
uname -m
```

Expected output:

```bash
x86_64
```

### 3.2 Download Official Syncthing Binary

```bash
mkdir -p ~/.local/bin
cd ~/.local/bin

ST_VER="v2.0.13"

curl -L -o syncthing.tar.gz \
  "https://github.com/syncthing/syncthing/releases/download/${ST_VER}/syncthing-linux-amd64-${ST_VER}.tar.gz"
```

### 3.3 Extract and Install

```bash
tar -xzf syncthing.tar.gz
cp syncthing-linux-amd64-${ST_VER}/syncthing .
rm -rf syncthing-linux-amd64-${ST_VER} syncthing.tar.gz
chmod +x syncthing
```

### 3.4 Verify Installation

```bash
which syncthing
syncthing --version
```

Expected output:

```bash
~/.local/bin/syncthing
syncthing v2.0.13 ...
```

---

## 4. Running Syncthing Persistently on NSCC (Required)

This is a **one-time setup**.

On NSCC, Syncthing runs as a normal user process.
If started in a regular SSH session, it will stop when the session ends.

To ensure Syncthing continues running in the background, it must be started
inside a `tmux` session.

---

### Step 1: Create a tmux session

```bash
tmux new -s syncthing
```

### Step 2: Start Syncthing inside tmux

```bash
syncthing --gui-address=127.0.0.1:8384
```

### Step 3: Detach from tmux (Syncthing keeps running)

Press:
Ctrl + B, then D

You can now safely close your SSH session. Syncthing will continue running in the background.

### Step 4: (Optional) Reattach later

```bash
tmux attach -t syncthing
```

### Step 5: (Optional) Stop Syncthing

```bash
pkill -u "$USER" syncthing
```

Notes
- This setup only needs to be done once
- Syncthing will automatically sync files whenever:
  - NSCC login node is available
  - Network connectivity exists
- This follows standard HPC best practices

---

## 5. Access Syncthing Web GUI (via SSH Tunnel)

NSCC does not expose ports publicly, so the GUI must be accessed via SSH port forwarding.
From your local laptop, open a new terminal:
```bash
ssh -N -L 8384:127.0.0.1:8384 <nscc_username>@aspire2antu.nscc.sg
```

Then open your browser and go to:
```bash
http://127.0.0.1:8384
```
---

## 6. Device Pairing (Host-Initiated)

**Important**  
Due to NSCC firewall restrictions, **all Syncthing device connections must be initiated from the Host Server GUI**.

Preferred: host-initiated. If no popup appears, manually add the Host device ID on NSCC.

1. On the **Host Server Syncthing GUI**:
   - Click **Add Remote Device**
   - Paste the **NSCC Device ID**
   - Save
  You can find the NSCC Device ID in the Syncthing GUI under:
  Actions → Show ID


2. On **NSCC Syncthing GUI**:
   - Accept the incoming device request

Device pairing is now complete.

---

## 7. Folder Sharing (Host-Defined)

**Important**  
- Folders must be **created and defined on the Host Server first**.  
- NSCC should only accept shared folders.

### Why `/sync` Does NOT Work

When Syncthing asks for a local folder path, using:

/sync

will NOT work.

This is because `/sync` is a **Docker mount point**, not a real filesystem
path visible to Syncthing.

### Correct Procedure (Host Server)

1. On the Host Server:
```bash
mkdir -p ./sync/<username>
```

2. In the Host Server Syncthing GUI, add a folder:
- Folder Label: <username>
- Folder Path: sync/<username>
- Folder Type: Send & Receive
- Share with: NSCC device

3. On NSCC:
- Accept the incoming folder share
- Do not change the folder path

---

## 8. Verify Sync

Use any experiment folder that has been shared from the Host Server.

On the Host Server:
```bash
ls ai-commons-watcher/sync/<username>/<experiment_folder>
```

Expected output:
Experiment files synced from NSCC

This confirms NSCC → Host Server syncing works.

---

## 9. Trigger Experiment Execution (Optional)

Inside the synced experiment folder on NSCC:
```bash
touch stop.txt
```

On the Host Server:
- Watcher detects the experiment
- Runs it automatically
- Removes stop.txt

---

## 10. Notes & Design Rationale

- Syncthing pairing and folder sharing are done via the Web GUI
- This is the officially supported Syncthing workflow
- CLI-only configuration is intentionally avoided for reliability
- Docker is required only on the Host Server, not NSCC

### Do I Need to Access the Syncthing GUI Again?
No.
After:
- Devices are paired
- Folder paths are configured
- Syncthing is running inside `tmux`

The Syncthing GUI is **not required** for normal operation.

The GUI is only needed for:
- Initial setup
- Debugging (rare)

---

## 11. Summary
This guide demonstrates that:
- NSCC users can run Syncthing without sudo
- Experiment data can be reliably synced to the Host Server
- The setup aligns with AI Commons’ Docker-first architecture

---
