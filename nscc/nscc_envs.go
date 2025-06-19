package nscc

import (
	"ai-commons/config"
	"strings"

	"golang.org/x/crypto/ssh"
)

func SetNSCCEnvironmentVariables(conn *ssh.Client) error {
	envs := config.GetConfig().NSCC.Envs
	if len(envs) == 0 {
		return nil // No environment variables to set
	}
	cmd := []string{}
	for _, env := range envs {
		if strings.TrimSpace(env.Key) == "" || strings.TrimSpace(env.Value) == "" {
			continue // Skip empty environment variables
		}
		cmd = append(cmd, "export "+env.Key+"=\""+env.Value+"\"")
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
