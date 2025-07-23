package utils

import (
	"ai-commons/config"
	"context"
	"fmt"
	"os"
)

func InitSSHKeys(ctx context.Context, hostname string, appendKnownHosts, appendSSHConfig bool) (map[string]string, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	accessToken := Getenv("BITWARDEN_ACCESS_TOKEN")
	if accessToken == "" {
		logger.Error("BITWARDEN_ACCESS_TOKEN environment variable is not set")
		return nil, fmt.Errorf("BITWARDEN_ACCESS_TOKEN environment variable is not set")
	}

	orgId := Getenv("BITWARDEN_ORG_ID")
	if orgId == "" {
		logger.Error("BITWARDEN_ORG_ID environment variable is not set")
		return nil, fmt.Errorf("BITWARDEN_ORG_ID environment variable is not set")
	}

	cfg := config.GetConfig()

	// Initialize Bitwarden client
	bwClient, err := GetBitwardenClient(
		accessToken,
		cfg.Bitwarden.ApiUrl,
		cfg.Bitwarden.IdentityUrl,
	)
	if err != nil {
		logger.Error("Failed to initialize Bitwarden client: ", err)
		panic(err)
	}

	defer CloseBitwardenClient(bwClient)

	// Download SSH keys from Bitwarden
	sshDir := cfg.SSH.KeysPath
	sshKeys, err := DownloadSSHKeys(ctx, bwClient, orgId, sshDir, true, cfg.ConfigDir, cfg.SSH.KeyPrefix)
	if err != nil {
		logger.Error("Failed to download SSH keys from Bitwarden: ", err)
		return nil, fmt.Errorf("failed to download SSH keys from Bitwarden: %v", err)
	}
	if len(sshKeys) == 0 {
		logger.Warn("No SSH keys were downloaded. Please check your Bitwarden vault and ensure there are SSH keys available.")
		return nil, fmt.Errorf("no SSH keys were downloaded")
	}

	if appendKnownHosts {
		// append known hosts
		err = AppendKnownHosts(ctx, hostname, cfg.SSH.KnownHostsPath)
		if err != nil {
			logger.Error("Failed to append known hosts: ", err)
			return nil, fmt.Errorf("failed to append known hosts: %v", err)
		}
	}

	if appendSSHConfig {
		sshConfigPath := cfg.SSH.ConfigPath

		// check if ssh config file exists, if yes, skip
		_, err := os.Stat(sshConfigPath)
		if err == nil {
			logger.Infof("SSH config file %s already exists, skipping creation", sshConfigPath)
		} else if !os.IsNotExist(err) {
			logger.Error("Failed to check SSH config file: ", err)
			return nil, fmt.Errorf("failed to check SSH config file: %v", err)
		}

		if os.IsNotExist(err) {
			if err := CreateDirFileIfNotExists(sshConfigPath); err != nil {
				logger.Error("Failed to create SSH config directory: ", err)
				return nil, fmt.Errorf("failed to create SSH config directory: %v", err)
			}
		}

		logger.Infof("Successfully created SSH config file %s", sshConfigPath)
		// append SSH keys to the SSH config file
		for key, path := range sshKeys {
			err = AppendSSHConfig(ctx, sshConfigPath, hostname, key, path)
			if err != nil {
				logger.Error("Failed to append SSH config: ", err)
				return nil, fmt.Errorf("failed to append SSH config: %v", err)
			}
			logger.Infof("Successfully appended SSH config to %s", sshConfigPath)
		}
	}

	return sshKeys, nil
}
