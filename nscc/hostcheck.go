package nscc

import (
	"ai-commons/utils"
	"context"
	"fmt"
	"strings"
)

func (node *Node) IsGitSetup(
	ctx context.Context,
) (bool, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return false, err
	}

	// Check if the connection is established
	if node.Conn == nil {
		return false, fmt.Errorf("SSH connection is nil")
	}

	// 1. Check if git is installed
	cmd := "git --version"
	out, _, err := utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		return false, fmt.Errorf("failed to run command: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		return false, fmt.Errorf("git is not installed on the host")
	}
	logger.Infof("Git version: %s", out)

	// 2. Check if username and email are configured
	cmd = "git config --get user.name"
	out, _, err = utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		return false, fmt.Errorf("failed to run command: %v. git user.name might not be configured on the host, %w", cmd, err)
	}
	logger.Infof("Git user.name: %s", out)

	cmd = "git config --get user.email"
	out, _, err = utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		return false, fmt.Errorf("failed to run command: %v. git user.email might not be configured on the host, %w", cmd, err)
	}
	logger.Infof("Git user.email: %s", out)
	logger.Info("Git is installed and configured on the host")

	// 3. Check if SSH folder is present with at least one key
	cmd = "ls -A ~/.ssh"
	out, _, err = utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	// if no folder doesn't exist, it will return an error
	if err != nil {
		return false, fmt.Errorf("failed to run command: %v", err)
	}
	// if the folder is empty, out will be an empty string
	if out == "" {
		return false, fmt.Errorf("SSH folder is empty on the host")
	}
	logger.Infof("SSH folder contents: %s", out)

	return true, nil
}
