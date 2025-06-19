package nscc

import (
	"ai-commons/config"
	"strings"

	"golang.org/x/crypto/ssh"
)

func AddNSCCModules(conn *ssh.Client) error {
	modules := config.GetConfig().NSCC.RequiredModules
	if len(modules) == 0 {
		return nil // No modules to add
	}
	cmd := []string{}
	for _, module := range modules {
		if strings.TrimSpace(module) == "" {
			continue // Skip empty module names
		}
		cmd = append(cmd, "module load "+module)
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run(strings.Join(cmd, " && ")); err != nil {
		return err
	}

	return nil
}
