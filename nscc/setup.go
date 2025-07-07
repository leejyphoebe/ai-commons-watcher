package nscc

import (
	"ai-commons/config"
	"ai-commons/utils"
	"context"
	"fmt"
	"strings"
)

func setupNode(ctx context.Context, exp config.ExperimentsConfig, node *Node) error {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return err
	}

	// Set environment variables
	if err := node.SetNSCCEnvironmentVariables(ctx); err != nil {
		logger.Errorf("Failed to set environment variables on node %s: %v", node.Host, err)
		return fmt.Errorf("failed to set environment variables on node %s: %w", node.Host, err)
	}

	if err := node.AddNSCCModules(ctx); err != nil {
		logger.Errorf("Failed to add NSCC modules on node %s: %v", node.Host, err)
		return fmt.Errorf("failed to add NSCC modules on node %s: %w", node.Host, err)
	}

	// rsync remote host
	user := node.Host
	hostname := config.GetConfig().SSH.Hostname
	host := fmt.Sprintf("%s@%s", user, hostname)

	// copy setup script to remote node
	remotePath := strings.Replace(exp.RemoteSetupScriptPath, "$USER", user, 1)
	dest := fmt.Sprintf("%s:%s", host, remotePath)
	err = utils.RsyncTransfer(ctx,
		node.Conn,
		exp.SetupScriptPath,
		dest,
		utils.RsyncLocalToRemote,
	)
	if err != nil {
		logger.Errorf("Failed to copy setup script to node %s: %v", node.Host, err)
		return err
	}
	logger.Infof("Successfully copied setup script to node %s", node.Host)

	// run setup script
	cmd := fmt.Sprintf("bash %s", exp.RemoteSetupScriptPath)
	out, _, err := utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		logger.Errorf("Failed to run setup script on node %s: %v", node.Host, err)
		return fmt.Errorf("failed to run setup script on node %s: %w", node.Host, err)
	}
	logger.Infof("Successfully ran setup script on node %s: %s", node.Host, out)

	return nil
}
