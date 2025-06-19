package main

import (
	"ai-commons/config"
	"ai-commons/nscc"
	"ai-commons/utils"
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

func main() {
	// Parse command line flags
	var (
		configFilePath string
	)

	flag.StringVar(&configFilePath, "config", "config.yaml", "Path to the configuration file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Load configuration
	err := config.InitConfig(configFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	cfg := config.GetConfig()

	// Initialize logger
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}

	logger := utils.GetBaseLogger().WithField("component", "main")
	logger.Info("Starting ai-commons...")

	// init SSH keys
	appendKnownHosts := true
	writeSSHConfig := true
	ctx := context.WithValue(context.Background(), utils.LoggerContextKey, logger)
	sshKeys, err := utils.InitSSHKeys(ctx, cfg.SSH.Hostname, appendKnownHosts, writeSSHConfig)
	if err != nil {
		logger.Error("Failed to initialize SSH keys: ", err)
		panic(err)
	}
	logger.Infof("Successfully initialized %d SSH keys", len(sshKeys))

	// connect to SSH host
	sshConns := make(map[string]*ssh.Client)
	logger.Info("Connecting to SSH hosts...")
	ctx = context.WithValue(ctx, utils.LoggerContextKey, logger)
	states := nscc.NodeStates{}
	for host := range sshKeys {
		conn, err := utils.GetConnection(ctx, host)
		if err != nil {
			logger.Errorf("Failed to connect to host %s: %v", host, err)
		}
		defer conn.Close()
		sshConns[host] = conn
		logger.Infof("Successfully connected to host %s", host)
		node := nscc.Node{Host: host, Conn: conn}
		state, err := node.GetNodeState(ctx)
		if err != nil {
			logger.Errorf("Failed to get node state for host %s: %v", host, err)
			continue
		}
		logger.Infof("Node state for host %s: %+v", host, node)
		states.Nodes[host] = state
	}

	yamlData, err := yaml.Marshal(states)
	if err != nil {
		logger.Errorf("Failed to marshal node state to YAML: %v", err)
		return
	}

	err = utils.WriteToFile(ctx, cfg.NodeStateFilePath, string(yamlData), 0644)
	if err != nil {
		logger.Errorf("Failed to write node state to file %s: %v", cfg.NodeStateFilePath, err)
		return
	}
	logger.Infof("Node state written to %s", cfg.NodeStateFilePath)
	logger.Infof("Node state YAML:\n%s", string(yamlData))
}
