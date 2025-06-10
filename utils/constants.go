package utils

import (
	"os"
	"path/filepath"
)

const (
	Hostname = "aspire2antu.nscc.sg"
)
var (
	CacheDir = filepath.Join(os.Getenv("PWD"), ".cache")
	// SSHConfigPath = filepath.Join(os.Getenv("HOME"), ".ssh", "nscc_config")
	// SSHKnownHostsPath = filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	SSHConfigPath = filepath.Join(CacheDir, "nscc_config")
	SSHKnownHostsPath = filepath.Join(CacheDir, "known_hosts")
)