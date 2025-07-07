package nscc

import (
	"ai-commons/config"
	"ai-commons/utils"
	"context"
	"strings"
)

func (node *Node) AddNSCCModules(ctx context.Context) error {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return err
	}

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
	cmd = append(cmd, "module list")

	err = utils.RunCommand(ctx, strings.Join(cmd, " && "), node.Conn)
	if err != nil {
		logger.Errorf("Failed to add NSCC modules: %v", err)
		return err
	}

	return nil
}
