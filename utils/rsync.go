package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// RsyncDirection defines the direction of the rsync transfer.
type RsyncDirection int

const (
	RsyncLocalToRemote  RsyncDirection = iota // Local machine to Remote server
	RsyncRemoteToLocal                        // Remote server to Local machine
	RsyncRemoteToRemote                       // Remote server to Remote server
)

// RsyncTransfer performs an rsync operation between specified paths,
// handling local-to-remote, remote-to-local, and remote-to-remote scenarios.
//
// ctx: Context for cancellation and logging.
// conn: The active SSH client connection to the remote server.
// src: The source path (local for L->R, remote for R->L, remote for R->R).
// dst: The destination path (remote for L->R, local for R->L, remote for R->R).
// direction: The direction of the rsync transfer.
// rsyncOptions: Optional rsync command-line flags (e.g., "-avz", "--delete").
func RsyncTransfer(ctx context.Context, conn *ssh.Client, src, dst string, direction RsyncDirection, rsyncOptions ...string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}

	// Default rsync options
	options := "-avz"
	if len(rsyncOptions) > 0 {
		options = strings.Join(rsyncOptions, " ")
	}

	switch direction {
	case RsyncRemoteToRemote:
		// This is the simplest case: rsync runs entirely on the remote server.
		logger.Infof("Initiating Remote-to-Remote rsync from '%s' to '%s'...", src, dst)
		return rsyncRemoteToRemote(ctx, conn, src, dst, options)

	case RsyncLocalToRemote:
		// For L->R, a local rsync client talks to a remote rsync --server process
		// over the SSH session's pipes.
		logger.Infof("Initiating Local-to-Remote rsync from local '%s' to remote '%s'...", src, dst)
		return rsyncLocalToRemote(ctx, src, dst, options)

	case RsyncRemoteToLocal:
		// For R->L, a local rsync client talks to a remote rsync --server --sender process
		// over the SSH session's pipes.
		logger.Infof("Initiating Remote-to-Local rsync from remote '%s' to local '%s'...", src, dst)
		return rsyncRemoteToLocal(ctx, conn, src, dst, options)

	default:
		return fmt.Errorf("unsupported rsync direction: %v", direction)
	}
}

// rsyncRemoteToRemote executes an rsync command directly on the remote server.
func rsyncRemoteToRemote(ctx context.Context, conn *ssh.Client, srcRemote, dstRemote, options string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}
	logger.Debugf("Starting Remote-to-Remote rsync from '%s' to '%s' with options: %s", srcRemote, dstRemote, options)

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session for remote-to-remote rsync: %w", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	quotedSrc := strconv.Quote(srcRemote)
	quotedDst := strconv.Quote(dstRemote)

	command := fmt.Sprintf("rsync %s %s %s", options, quotedSrc, quotedDst)
	logger.Debugf("Executing remote rsync command: %s", command)

	if err := session.Run(command); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			return fmt.Errorf("remote rsync command failed with exit status %d: %w", exitErr.ExitStatus(), err)
		}
		return fmt.Errorf("failed to run remote rsync command: %w", err)
	}

	logger.Infof("Remote-to-Remote rsync completed successfully: '%s' -> '%s'", srcRemote, dstRemote)
	return nil
}

// rsyncLocalToRemote transfers files from local to remote using rsync over SSH pipes.
func rsyncLocalToRemote(ctx context.Context, srcLocal, dstRemote, options string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}
	logger.Debugf("Starting Local-to-Remote rsync from '%s' to '%s' with options: %s", srcLocal, dstRemote, options)

	args := fmt.Sprintf("%s -e ssh %s %s", options, srcLocal, dstRemote)
	logger.Infof("Executing local to remote rsync command: rsync %s", args)
	cmd := exec.Command("rsync", strings.Split(args, " ")...)

	cmd.Stderr = os.Stderr

	var wg sync.WaitGroup
	wg.Add(1)
	var localErr error

	// Run the local rsync client
	go func() {
		defer wg.Done()
		out, err := cmd.Output()
		if err != nil {
			logger.Errorf("Running local to remote rsync failed: %v", err)
		}
		logger.Infof("Rsync output: %s", out)
	}()

	wg.Wait()

	if localErr != nil {
		return fmt.Errorf("Local-to-Remote rsync failed: %w", localErr)
	}

	logger.Infof("Local-to-Remote rsync completed successfully: '%s' -> '%s'", srcLocal, dstRemote)
	return nil
}

// rsyncRemoteToLocal transfers files from remote to local using rsync over SSH pipes.
func rsyncRemoteToLocal(ctx context.Context, conn *ssh.Client, srcRemote, dstLocal, options string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get logger from context: %w", err)
	}
	logger.Debugf("Starting Remote-to-Local rsync from '%s' to '%s' with options: %s", srcRemote, dstLocal, options)
	// 1. Prepare the remote rsync server command (acting as sender)
	remoteSession, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session for R->L remote rsync server: %w", err)
	}
	defer remoteSession.Close()

	remoteStdin, err := remoteSession.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe for remote rsync server: %w", err)
	}
	remoteStdout, err := remoteSession.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe for remote rsync server: %w", err)
	}
	remoteSession.Stderr = os.Stderr

	// Start remote rsync --server --sender (it sends to the client)
	serverCommand := fmt.Sprintf("rsync --server --sender %s %s", options, strconv.Quote(srcRemote))
	logger.Debugf("Starting remote rsync server (R->L): %s", serverCommand)

	if err := remoteSession.Start(serverCommand); err != nil {
		return fmt.Errorf("failed to start remote rsync server for R->L: %w", err)
	}

	// 2. Prepare the local rsync client command (acting as receiver)
	localRsyncArgs := []string{
		options,
		".",      // Remote source is implicit via the remote --server command
		dstLocal, // Local destination path
		"--rsh=/bin/sh",
		"--numeric-ids",
		"--inplace",
	}
	// The client communicates with --server --sender
	localRsyncArgs = append([]string{"--client"}, localRsyncArgs...)
	localRsyncCmd := exec.Command("rsync", localRsyncArgs...)
	logger.Debugf("Starting local rsync client (R->L): rsync %v", localRsyncArgs)

	// Connect local rsync's stdout to remote session's stdin
	localRsyncCmd.Stdout = remoteStdin
	// Connect remote session's stdout to local rsync's stdin
	localRsyncCmd.Stdin = remoteStdout
	// Local rsync's stderr should go to our local stderr
	localRsyncCmd.Stderr = os.Stderr

	var wg sync.WaitGroup
	wg.Add(2)
	var localErr, remoteErr error

	// Run the local rsync client
	go func() {
		defer wg.Done()
		localErr = localRsyncCmd.Run()
		if localErr != nil {
			logger.Errorf("Local rsync client (R->L) failed: %v", localErr)
		}
		// Closing the remoteStdin pipe signals EOF to the remote rsync server.
		remoteStdin.Close()
	}()

	// Wait for the remote rsync server to exit
	go func() {
		defer wg.Done()
		remoteErr = remoteSession.Wait()
		if remoteErr != nil {
			logger.Errorf("Remote rsync server (R->L) failed: %v", remoteErr)
		}
	}()

	wg.Wait()

	if localErr != nil {
		return fmt.Errorf("Remote-to-Local rsync failed on client side: %w", localErr)
	}
	if remoteErr != nil {
		return fmt.Errorf("Remote-to-Local rsync failed on remote server side: %w", remoteErr)
	}

	logger.Infof("Remote-to-Local rsync completed successfully: '%s' -> '%s'", srcRemote, dstLocal)
	return nil
}
