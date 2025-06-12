package main

import (
	"ai-commons/config"
	"ai-commons/utils"
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
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
	for host := range sshKeys {
		conn, err := utils.GetConnection(ctx, host)
		if err != nil {
			logger.Errorf("Failed to connect to host %s: %v", host, err)
			panic(err)
		}
		defer conn.Close()
		sshConns[host] = conn
		logger.Infof("Successfully connected to host %s", host)
		// check credit + number of gpus currently used
		utils.RunCommand(ctx, "echo '\nHello World from "+host+"!\n'", conn)
	}

	// check if git is setup in login node
	ok, err := utils.IsGitSetup(ctx, sshConns[cfg.SSH.MasterHost])
	if err != nil {
		logger.Errorf("Failed to check git setup: %v", err)
		panic(err)
	}
	if !ok {
		logger.Errorf("git is not set up correctly for %s. Please ensure git is installed, user.name and user.email are configured, and SSH keys are present", cfg.SSH.MasterHost)
		os.Exit(1)
	}

	// handle signals for graceful shutdown


	// setup environment variables in login node


	// install modules in login node


	// run submit job commands in login node


	// cleanup


	// write tests

}
	