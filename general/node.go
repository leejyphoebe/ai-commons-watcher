package general

import (
	"ai-commons/utils"

	"golang.org/x/crypto/ssh"
)

// Note: not NSCC specific
type Node struct {
	Host    string
	Conn    *ssh.Client
	Cluster utils.Cluster // Cluster name derived from hostname
}
