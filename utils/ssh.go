package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func hostKeyCallback() (ssh.HostKeyCallback, error) {
	knownHostsPath := SSHKnownHostsPath
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("could not create known_hosts callback: %w", err)
	}
	return callback, nil
}

func LoadSSHConfig(ctx context.Context, host string) (*ssh.ClientConfig, error) {
	// Get the logger from the context
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Load the SSH configuration from the specified path
	logger.Infof("Loading SSH configuration from %s", SSHConfigPath)
	file, err := os.Open(SSHConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH config file %s: %w", SSHConfigPath, err)
	}
	defer file.Close()

	sshConfig, err := ssh_config.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SSH config file %s: %w", SSHConfigPath, err)
	}

	user, err := sshConfig.Get(host, "User")
	if err != nil {
		return nil, fmt.Errorf("failed to get User from host %s in SSH config file %s: %w", host, SSHConfigPath, err)
	}

	keyPath, err := sshConfig.Get(host, "IdentityFile")
	if err != nil {
		return nil, fmt.Errorf("failed to get IdentityFile from host %s in SSH config file %s: %w", host, SSHConfigPath, err)
	}
	// Load the private key from the specified path
	logger.Infof("Loading private key from %s for user %s", keyPath, user)
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from %s: %w", keyPath, err)
	}
	// Parse the private key
	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key from %s: %w", keyPath, err)
	}

	hostKeyCallback, err := hostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("failed to create host key callback: %w", err)
	}

	// Create a new SSH client configuration
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout: 5 * time.Second,
	}

	return clientConfig, nil
}

func GetConnection(ctx context.Context, host string) (*ssh.Client, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	clientConfig, err := LoadSSHConfig(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH config for host %s: %w", host, err)
	}

	logger.Debugf("Connecting to %s", host)
	conn, err := ssh.Dial("tcp", Hostname+":22", clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", host, err)
	}

	logger.Infof("Successfully connected to %s", host)
	return conn, nil
}

func RunCommand(cmd string, conn *ssh.Client) error {
	sess, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer sess.Close()

	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get SSH session stdout pipe: %v", err)
	}
	go io.Copy(os.Stdout, sessStdOut)

	sessStderr, err := sess.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get SSH session stderr pipe: %v", err)
	}
	go io.Copy(os.Stderr, sessStderr)
	err = sess.Run(cmd) // eg., /usr/bin/whoami
	if err != nil {
		return fmt.Errorf("failed to run command %q: %v", cmd, err)
	}
	return nil
}