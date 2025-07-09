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

	// rsync remote host
	user := node.Host
	hostname := config.GetConfig().SSH.Hostname
	host := fmt.Sprintf("%s@%s", user, hostname)

	// create setup dir if not exists
	setupDir := strings.Replace(exp.ExperimentSetup.SetupFilesDir, "$USER", user, 1)
	cmd := fmt.Sprintf("mkdir -p %s", setupDir)
	out, _, err := utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		logger.Errorf("Failed to create setup directory on node %s: %v", node.Host, err)
		return fmt.Errorf("failed to create setup directory on node %s: %w", node.Host, err)
	}
	logger.Infof("Successfully created setup directory on node %s: %s", node.Host, out)

	// copy setup files to remote node
	for _, file := range exp.ExperimentSetup.SetupFiles {
		src := file.Src
		dest := fmt.Sprintf("%s:%s/%s", host, setupDir, file.Dest)
		err = utils.RsyncTransfer(ctx, node.Conn, src, dest, utils.RsyncLocalToRemote)
		if err != nil {
			logger.Errorf("Failed to copy setup file %s to node %s: %v", src, node.Host, err)
			return err
		}
		logger.Infof("Successfully copied setup file %s to node %s", src, node.Host)
	}

	// copy setup script to remote node
	remoteScriptPath := fmt.Sprintf("%s/%s", setupDir, exp.ExperimentSetup.SetupScriptRemotePath)
	dest := fmt.Sprintf("%s:%s", host, remoteScriptPath)
	err = utils.RsyncTransfer(ctx,
		node.Conn,
		exp.ExperimentSetup.SetupScriptLocalPath,
		dest,
		utils.RsyncLocalToRemote,
	)
	if err != nil {
		logger.Errorf("Failed to copy setup script to node %s: %v", node.Host, err)
		return err
	}
	logger.Infof("Successfully copied setup script to node %s", node.Host)

	// run setup script
	cmd = fmt.Sprintf("bash -e %s", remoteScriptPath)
	out, _, err = utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		logger.Errorf("Failed to run setup script on node %s: %v", node.Host, err)
		return fmt.Errorf("failed to run setup script on node %s: %w", node.Host, err)
	}
	logger.Infof("Successfully ran setup script on node %s: %s", node.Host, out)

	return nil
}
