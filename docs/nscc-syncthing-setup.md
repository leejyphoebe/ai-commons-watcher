# NSCC Syncthing Setup Guide (User-Level, No Sudo)

This guide explains how to set up **Syncthing on NSCC** to sync experiment data to the host server.

## 1. Prerequisites

- NSCC account with SSH access
- Working internet connection on NSCC
- Host Server Syncthing already running (Dockerised)

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

## 6. Pair with Host Server (GUI)

### 6.1 Add Host Server as Remote Device (NSCC side)

In the NSCC Syncthing GUI:
1. Click Add Remote Device
2. Paste the Host Server Device ID
3. Set a recognisable name (e.g. ai-commons-host)
4. Click Save

### 6.2 Accept Device on Host Server

On the Host Server Syncthing GUI:
1. Accept the incoming device request
Pairing is now complete.

---

## 7. Create and Share an Experiment Folder

### 7.1 Create Folder on NSCC (CLI)

```bash
mkdir -p ~/experiments/<username>/test_exp_demo
echo "hello ai commons" > ~/experiments/<username>/test_exp_demo/README.txt
```

### 7.2 Add Folder in NSCC GUI
In the NSCC Syncthing GUI:
- Click Add Folder

Folder Label
```bash
test_exp_demo
```

Folder Path
```bash
/home/users/ntu/<username>/experiments/<username>/test_exp_demo
```

Folder Type
```bash
Send & Receive
```

Under Sharing:
- Tick the Host Server
- Click Save

### 7.3 Accept Folder on Host Server
On the Host Server Syncthing GUI:
1. Accept the folder
2. Leave the default path unchanged
3. Click Save

---

## 8. Verify Sync

On the Host Server:
```bash
ls /home/<host_user>/ai-commons-watcher/sync/<username>/test_exp_demo
```

Expected output:
```bash
README.txt
```
This confirms NSCC → Host Server syncing works.

---

## 9. Trigger Experiment Execution (Optional)

To trigger the watcher, on NSCC:
```bash
touch ~/experiments/<username>/test_exp_demo/stop.txt
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

---

## 11. Summary
This guide demonstrates that:
- NSCC users can run Syncthing without sudo
- Experiment data can be reliably synced to the Host Server
- The setup aligns with AI Commons’ Docker-first architecture

---
