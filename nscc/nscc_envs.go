package nscc

import (
	"ai-commons/config"
	"ai-commons/utils"
	"context"
	"strings"
)

func (node *Node) SetNSCCEnvironmentVariables(ctx context.Context) error {
	// Get the logger from the context
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return err
	}

	envs := config.GetConfig().NSCC.Envs
	if len(envs) == 0 {
		return nil // No environment variables to set
	}
	cmd := []string{}
	keys := []string{}
	for _, env := range envs {
		if strings.TrimSpace(env.Key) == "" || strings.TrimSpace(env.Value) == "" {
			continue // Skip empty environment variables
		}
		cmd = append(cmd, "export "+env.Key+"=\""+env.Value+"\"")
		keys = append(keys, env.Key)
	}
	cmd = append(cmd, "printenv | grep -E '("+strings.Join(keys, "|")+")'")

	err = utils.RunCommand(ctx, strings.Join(cmd, " && "), node.Conn)
	if err != nil {
		logger.Errorf("Failed to set NSCC environment variables: %v", err)
		return err
	}
	logger.Info("NSCC environment variables set successfully")

	return nil
}
