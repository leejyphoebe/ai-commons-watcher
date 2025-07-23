package utils

import (
	"ai-commons/config"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kevinburke/ssh_config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func hostKeyCallback() (ssh.HostKeyCallback, error) {
	knownHostsPath := config.GetConfig().SSH.KnownHostsPath
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("could not create known_hosts callback: %w", err)
	}
	return callback, nil
}

func LoadSSHConfig(ctx context.Context, host string, timeout int) (*ssh.ClientConfig, error) {
	// Get the logger from the context
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Load the SSH configuration from the specified path
	sshConfigPath := config.GetConfig().SSH.ConfigPath
	logger.Infof("Loading SSH configuration from %s", sshConfigPath)
	file, err := os.Open(sshConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH config file %s: %w", sshConfigPath, err)
	}
	defer file.Close()

	sshConfig, err := ssh_config.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SSH config file %s: %w", sshConfigPath, err)
	}

	user, err := sshConfig.Get(host, "User")
	if err != nil {
		return nil, fmt.Errorf("failed to get User from host %s in SSH config file %s: %w", host, sshConfigPath, err)
	}

	keyPath, err := sshConfig.Get(host, "IdentityFile")
	if err != nil {
		return nil, fmt.Errorf("failed to get IdentityFile from host %s in SSH config file %s: %w", host, sshConfigPath, err)
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
		Timeout:         time.Duration(timeout) * time.Second,
	}

	return clientConfig, nil
}

func GetConnection(ctx context.Context, host string) (*ssh.Client, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	attempts := 0
	cfg := config.GetConfig()
	sleepDuration := cfg.SSH.SleepSeconds
	timeout := cfg.SSH.TimeoutSeconds
	maxAttempts := cfg.SSH.MaxAttempts
	logger = logger.WithField("host", host)
	logger.Infof("Attempting to connect to %s with max attempts %d", host, maxAttempts)
	for {
		attempts++
		if attempts > maxAttempts {
			return nil, fmt.Errorf("failed to connect to %s after %d attempts", host, maxAttempts)
		}
		logger.Infof("Attempting to connect to %s (attempt %d/%d)", host, attempts, maxAttempts)

		clientConfig, err := LoadSSHConfig(ctx, host, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to load SSH config for host %s: %w", host, err)
		}

		logger.Debugf("Connecting to %s", host)
		conn, err := ssh.Dial("tcp", config.GetConfig().SSH.Hostname+":22", clientConfig)
		if err != nil {
			logger.Warnf("Failed to connect to %s: %v", host, err)
			time.Sleep(time.Duration(sleepDuration) * time.Second)
			continue
		} else {
			logger.Infof("Successfully connected to %s", host)
			return conn, nil
		}
	}
}

func RunCommand(ctx context.Context, cmd string, conn *ssh.Client) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}
	logger.WithFields(log.Fields{
		"command": cmd,
		"host":    conn.Conn.RemoteAddr().String(),
	})
	logger.Debugf("Running command: %s", cmd)

	// Create a new SSH session
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

func RunCommandGetOutput(ctx context.Context, cmd string, conn *ssh.Client) (string, string, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve logger from context: %v", err)
	}
	logger.WithFields(log.Fields{
		"command": cmd,
		"host":    conn.RemoteAddr().String(),
	})
	logger.Debugf("Running command: %s", cmd)
	sess, err := conn.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer sess.Close()

	// create pipe for stdout and stderr
	stdoutPipe, err := sess.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to get SSH session stdout pipe: %v", err)
	}
	stderrPipe, err := sess.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to get SSH session stderr pipe: %v", err)
	}
	// create buffer to capture output
	var stdoutBuf, stderrBuf bytes.Buffer

	multiStdoutWriter := io.MultiWriter(os.Stdout, &stdoutBuf)
	multiStderrWriter := io.MultiWriter(os.Stderr, &stderrBuf)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := io.Copy(multiStdoutWriter, stdoutPipe)
		if err != nil && err != io.EOF {
			logger.Errorf("failed to copy stdout: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		_, err := io.Copy(multiStderrWriter, stderrPipe)
		if err != nil && err != io.EOF {
			logger.Errorf("failed to copy stderr: %v", err)
		}
	}()

	// run the command
	runErr := sess.Start(cmd)
	if runErr != nil {
		// Wait for goroutines to finish copying any buffered output before returning
		wg.Wait()
		logger.WithError(runErr).Errorf("Failed to start command '%s'. Stderr: %s", cmd, stderrBuf.String())
		return "", "", fmt.Errorf("failed to start command '%s': %w", cmd, runErr)
	}

	// Wait for the command to complete
	runErr = sess.Wait()

	// IMPORTANT: Wait for the io.Copy goroutines to finish AFTER sess.Wait()
	// This ensures all output has been fully streamed and captured.
	wg.Wait()

	if runErr != nil {
		if exitErr, ok := runErr.(*ssh.ExitError); ok {
			logger.WithError(exitErr).WithField("exit_code", exitErr.ExitStatus()).
				Errorf("Command '%s' exited with non-zero status. Stderr: %s", cmd, stderrBuf.String())
			return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("command '%s' exited with status %d: %s",
				cmd, exitErr.ExitStatus(), strings.TrimSpace(stderrBuf.String()))
		}
		logger.WithError(runErr).Errorf("Error waiting for command '%s' to complete. Stderr: %s", cmd, stderrBuf.String())
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("error waiting for command '%s': %w, Stderr: %s",
			cmd, runErr, strings.TrimSpace(stderrBuf.String()))
	}

	logger.Debugf("Command '%s' executed successfully. Stdout: '%s', Stderr: '%s'",
		cmd, strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()))

	return stdoutBuf.String(), stderrBuf.String(), nil
}
