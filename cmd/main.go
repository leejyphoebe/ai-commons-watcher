package main

import (
	"ai-commons/utils"
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)



func main() {
	logFile := ""
	logLevel := log.InfoLevel
	outputJson := false
	logToStdout := true

	if err := utils.InitLogger(logFile, logLevel, outputJson, logToStdout); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}

	mainLogger := utils.GetBaseLogger().WithField("component", "main")
	mainLogger.Info("Starting ai-commons...")

	// init SSH keys
	appendKnownHosts := true
	writeSSHConfig := true
	sshInitLogger := utils.GetBaseLogger().WithField("component", "ssh_init")
	ctx := context.WithValue(context.Background(), utils.LoggerContextKey, sshInitLogger)
	sshKeys, err := utils.InitSSHKeys(ctx, utils.Hostname, appendKnownHosts, writeSSHConfig)
	if err != nil {
		sshInitLogger.Error("Failed to initialize SSH keys: ", err)
		panic(err)
	}
	sshInitLogger.Infof("Successfully initialized %d SSH keys", len(sshKeys))

	// connect to SSH host
	sshConns := make(map[string]*ssh.Client)
	sshConnectLogger := utils.GetBaseLogger().WithField("component", "ssh_connect")
	sshConnectLogger.Info("Connecting to SSH hosts...")
	ctx = context.WithValue(ctx, utils.LoggerContextKey, sshConnectLogger)
	for host := range sshKeys {
		conn, err := utils.GetConnection(ctx, host)
		if err != nil {
			sshInitLogger.Errorf("Failed to connect to host %s: %v", host, err)
			panic(err)
		}
		defer conn.Close()
		sshConns[host] = conn
		sshInitLogger.Infof("Successfully connected to host %s", host)
		utils.RunCommand("echo '\nHello World from "+host+"!\n'", conn)
	}
}
