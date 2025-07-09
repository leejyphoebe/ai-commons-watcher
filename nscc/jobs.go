package nscc

import (
	"ai-commons/config"
	"ai-commons/utils"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func CombineJobHosts() []string {
	hosts := make([]string, 0)
	for _, exp := range config.GetConfig().Experiments {
		for _, cfgNode := range exp.Nodes {
			hosts = append(hosts, cfgNode.Host)
		}
	}
	return hosts
}

func RunJobs(ctx context.Context) (string, string, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get logger from context: %w", err)
	}

	// ssh into login nodes
	selectedHosts := CombineJobHosts()
	sshConns := make(map[string]*ssh.Client)
	logger.Info("Connecting to SSH hosts...")
	states := NodeStates{}
	states.Nodes = make(map[string]NodeState)
	for _, host := range selectedHosts {
		conn, err := utils.GetConnection(ctx, host)
		if err != nil {
			logger.Errorf("Failed to connect to host %s: %v", host, err)
		}
		defer conn.Close()
		sshConns[host] = conn
		logger.Infof("Successfully connected to host %s", host)
		node := Node{Host: host, Conn: conn}
		state, err := node.GetNodeState(ctx)
		if err != nil {
			logger.Errorf("Failed to get node state for host %s: %v", host, err)
			continue
		}
		states.Nodes[host] = state
		logger.Infof("Node state for host %s: %+v", host, node)
	}

	// check if it is possible to run the job
	for host, state := range states.Nodes {
		if !state.CanConnect || !state.IsReachable {
			logger.Errorf("Node %s is not reachable or cannot connect", host)
			return "", "", fmt.Errorf("node %s is not reachable or cannot connect", host)
		}
		if !state.IsGitSetup {
			logger.Errorf("Git is not set up on node %s", host)
			return "", "", fmt.Errorf("git is not set up on node %s", host)
		}
	}

	// copy and run setup script on remote nodes
	for _, exp := range config.GetConfig().Experiments {
		for _, cfgNode := range exp.Nodes {
			node := Node{Host: cfgNode.Host, Conn: sshConns[cfgNode.Host]}
			err := setupNode(ctx, exp, &node)
			if err != nil {
				logger.Errorf("Failed to set up node %s: %v", node.Host, err)
				return "", "", fmt.Errorf("failed to set up node %s: %w", node.Host, err)
			}
			user := node.Host
			hostname := config.GetConfig().SSH.Hostname
			host := fmt.Sprintf("%s@%s", user, hostname)

			// copy experiment configs to remote node
			for _, expConfig := range cfgNode.ExperimentConfigPaths {
				absPath, err := filepath.Abs(exp.LocalConfigDir)
				if err != nil {
					logger.Errorf("Failed to get absolute path for local config dir %s: %v", exp.LocalConfigDir, err)
					return "", "", fmt.Errorf("failed to get absolute path for local config dir %s: %w", exp.LocalConfigDir, err)
				}
				remotePath := strings.Replace(exp.RemoteConfigDir, "$USER", user, 1)

				src := filepath.Join(absPath, expConfig.Src)
				dest := fmt.Sprintf("%s:%s/%s", host, remotePath, expConfig.Dest)
				if src == "" || dest == "" {
					logger.Errorf("Experiment config paths are not properly defined for node %s", node.Host)
					return "", "", fmt.Errorf("experiment config paths are not properly defined for node %s", node.Host)
				}

				err = utils.RsyncTransfer(ctx,
					node.Conn,
					src,
					dest,
					utils.RsyncLocalToRemote,
				)

				if err != nil {
					logger.Errorf("Failed to copy experiment config from %s to %s on node %s: %v", src, dest, node.Host, err)
					return "", "", err
				}
				logger.Infof("Successfully copied experiment config from %s to %s on node %s", src, dest, node.Host)

				// run command to run the experiment
				cmdDir := strings.Replace(exp.CmdDir, "$USER", user, 1)
				cmd := fmt.Sprintf("cd %s && %s", cmdDir, expConfig.Command)
				if cmd == "" {
					logger.Errorf("Command is not defined for experiment config %s on node %s", expConfig.Src, node.Host)
					return "", "", fmt.Errorf("command is not defined for experiment config %s on node %s", expConfig.Src, node.Host)
				}
				out, _, err := utils.RunCommandGetOutput(ctx, cmd, node.Conn)
				if err != nil {
					logger.Errorf("Failed to run command '%s' on node %s: %v", cmd, node.Host, err)
					return "", "", fmt.Errorf("failed to run command '%s' on node %s: %w", cmd, node.Host, err)
				}
				logger.Infof("Successfully ran command '%s' on node %s: %s", cmd, node.Host, out)
				// Log the output of the command
				if out != "" {
					logger.Infof("Output from command '%s' on node %s: %s", cmd, node.Host, out)
				} else {
					logger.Infof("No output from command '%s' on node %s", cmd, node.Host)
				}
			}
		}
	}

	// run command
	return "", "", nil
}
