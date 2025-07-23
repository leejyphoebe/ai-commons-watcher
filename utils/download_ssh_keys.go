package utils

import (
	"ai-commons/config"
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitwarden/sdk-go"
)

func GetBitwardenClient(accessToken, apiUrl, identityUrl string) (*sdk.BitwardenClientInterface, error) {
	if apiUrl == "" {
		apiUrl = "https://api.bitwarden.eu"
	}
	if identityUrl == "" {
		identityUrl = "https://identity.bitwarden.eu"
	}

	// Create a new Bitwarden client with the provided Urls
	bitwardenClient, err := sdk.NewBitwardenClient(&apiUrl, &identityUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to create Bitwarden client: %v", err)
	}

	// Login to Bitwarden
	stateFile := config.GetConfig().Bitwarden.StateFile
	// Attempt to login using the access token
	err = bitwardenClient.AccessTokenLogin(accessToken, &stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to login to Bitwarden with access token: %v", err)
	}

	return &bitwardenClient, nil
}

// Close the Bitwarden client when done
func CloseBitwardenClient(client *sdk.BitwardenClientInterface) {
	if client != nil {
		(*client).Close()
	}
}

func readSecretIdsFromCache(ctx context.Context, cacheFilePath string) (map[string]string, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	sshKeys := make(map[string]string)

	file, err := os.Open(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Cache file does not exist, not an error for this function
		}
		return nil, fmt.Errorf("failed to open cache file %s: %v", cacheFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ",", 2) // Split only on the first comma
		if len(parts) == 2 {
			sshKeys[parts[0]] = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading cache file %s: %v", cacheFilePath, err)
	}

	if len(sshKeys) == 0 {
		return nil, fmt.Errorf("cache file %s is empty or malformed", cacheFilePath)
	}

	logger.Infof("Bitwarden secrets list loaded from cache: %s\n", cacheFilePath)
	return sshKeys, nil
}

func writeSecretIdsToCache(ctx context.Context, cacheFilePath string, sshKeys map[string]string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	file, err := os.Create(cacheFilePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file %s: %v", cacheFilePath, err)
	}
	defer file.Close()

	for key, id := range sshKeys {
		_, err := fmt.Fprintf(file, "%s,%s\n", key, id)
		if err != nil {
			return fmt.Errorf("failed to write to cache file %s: %v", cacheFilePath, err)
		}
	}

	logger.Infof("Bitwarden secrets list cached to %s\n", cacheFilePath)
	return nil
}

// list secret ids and keys starting with ssh key prefix
func ListSSHKeys(ctx context.Context, client *sdk.BitwardenClientInterface, orgId, cacheDir, keyPrefix string, cache bool) (map[string]string, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	sshKeys := make(map[string]string)

	if cache {
		cachedSecrets, err := readSecretIdsFromCache(ctx, fmt.Sprintf("%s/ssh_keys.csv", cacheDir))
		if err == nil && cachedSecrets != nil {
			logger.Infof("Using cached secret ids from %s\n", fmt.Sprintf("%s/ssh_keys.csv", cacheDir))
			return cachedSecrets, nil
		}
		logger.Infof("Cache read failed or cache is empty (%v), fetching from Bitwarden...\n", err)
	}

	// Get all secrets
	secrets, err := (*client).Secrets().List(orgId)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %v", err)
	}

	for _, secret := range secrets.Data {
		if secret.Key != "" && len(secret.Key) > len(keyPrefix)+1 && strings.HasPrefix(secret.Key, keyPrefix) {
			sshKeys[secret.Key] = secret.ID
		}
	}

	if len(sshKeys) == 0 {
		return nil, fmt.Errorf("no SSH keys found")
	}

	if cache {
		err = writeSecretIdsToCache(ctx, fmt.Sprintf("%s/ssh_keys.csv", cacheDir), sshKeys)
		if err != nil {
			return nil, fmt.Errorf("failed to write SSH keys to cache: %v", err)
		}
	}

	return sshKeys, nil
}

// GetSSHKey retrieves the value of a specific SSH key by its ID
func GetSSHKey(client *sdk.BitwardenClientInterface, secretId string) (string, error) {
	secret, err := (*client).Secrets().Get(secretId)
	if err != nil {
		return "", fmt.Errorf("failed to get secret with ID %s: %v", secretId, err)
	}

	if secret == nil || secret.Value == "" {
		return "", fmt.Errorf("SSH key with ID %s not found or has no value", secretId)
	}

	return secret.Value, nil
}

func CheckIfSSHKeyExists(saveDir, keyPrefix, key string) (bool, error) {
	// Construct the file path for the SSH key
	filePath := fmt.Sprintf("%s/%s%s", saveDir, keyPrefix, key[8:]) // Remove "ssh_key_" prefix

	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil // File does not exist
	}
	if err != nil {
		return false, fmt.Errorf("error checking if SSH key exists: %v", err)
	}

	return true, nil // File exists
}

// future use case for adding more SSH keys
func CheckMissingSSHKeys(saveDir, keyPrefix string, sshKeys map[string]string) ([]string, error) {
	missingKeys := []string{}

	for key := range sshKeys {
		exists, err := CheckIfSSHKeyExists(saveDir, keyPrefix, key)
		if err != nil {
			return nil, fmt.Errorf("error checking if SSH key %s exists: %v", key, err)
		}
		if !exists {
			missingKeys = append(missingKeys, key)
		}
	}

	return missingKeys, nil
}

func DownloadSSHKey(ctx context.Context, client *sdk.BitwardenClientInterface, keyName, secretId, saveDir, keyPrefix string) (string, error) {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Construct the file path for the SSH key
	filePath := fmt.Sprintf("%s/%s%s", saveDir, keyPrefix, keyName[8:]) // Remove "ssh_key_" prefix
	// check if the SSH key already exists
	exists, err := CheckIfSSHKeyExists(saveDir, keyPrefix, keyName)
	if err != nil {
		return "", fmt.Errorf("error checking if SSH key %s exists: %v", keyName, err)
	}
	if exists {
		logger.Debugf("SSH key %s already exists in %s, skipping download\n", keyName, filePath)
		return filePath, nil // Skip download if the key already exists
	}

	// Get the SSH key value from Bitwarden
	if keyName == "" || secretId == "" {
		return "", fmt.Errorf("key or secret ID is empty for SSH key %s", keyName)
	}

	sshKey, err := GetSSHKey(client, secretId)
	if err != nil {
		return "", fmt.Errorf("failed to get SSH key %s: %v", keyName, err)
	}

	err = os.WriteFile(filePath, []byte(sshKey), 0600) // Set permissions to 0600
	if err != nil {
		return "", fmt.Errorf("failed to write SSH key to file %s: %v", filePath, err)
	}
	logger.Infof("SSH key %s downloaded successfully to %s\n", keyName, filePath)
	return filePath, nil
}

// DownloadSSHKeys: downloads all SSH keys from Bitwarden and saves them to the specified directory
// TODO:
// - add timeout for Bitwarden API requests
// - add error handling for file writing
// - add exponential backoff for querying Bitwarden API
func DownloadSSHKeys(ctx context.Context, client *sdk.BitwardenClientInterface, orgId, saveDir string, cache bool, cacheDir, keyPrefix string) (map[string]string, error) {
	// Validate input parameters
	if client == nil {
		return nil, fmt.Errorf("bitwarden client is nil")
	}
	if orgId == "" {
		return nil, fmt.Errorf("organization ID is empty")
	}

	// Check if save directory exists, if not return an error
	exists, err := CheckIfDirExists(saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check if save directory exists: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("save directory %s does not exist", saveDir)
	}

	// Check if cache directory exists, if not create it
	if cache {
		err := CreateDirFileIfNotExists(fmt.Sprintf("%s/ssh_keys.csv", cacheDir))
		if err != nil {
			return nil, fmt.Errorf("failed to create cache file %s: %v", cacheDir, err)
		}
	}

	// list SSH keys from Bitwarden
	sshKeys, err := ListSSHKeys(ctx, client, orgId, cacheDir, keyPrefix, cache)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %v", err)
	}

	if len(sshKeys) == 0 {
		return nil, fmt.Errorf("no SSH keys found in organization %s", orgId)
	}

	// Download each SSH key and save it to the specified directory
	sshKeysMap := make(map[string]string)
	for key, id := range sshKeys {
		sshKeyLocation, err := DownloadSSHKey(ctx, client, key, id, saveDir, keyPrefix)
		if err != nil {
			return nil, fmt.Errorf("failed to download SSH key %s: %v", key, err)
		}
		sshKeysMap[key[8:]] = sshKeyLocation
	}

	return sshKeysMap, nil
}

func AppendKnownHosts(ctx context.Context, hostname, knownHostsPath string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// check in cache to see if known_hosts has been appended to before
	if _, err := os.Stat(knownHostsPath); err == nil {
		logger.Infof("known_hosts cache %s already exists, skipping append", knownHostsPath)
		return nil
	}

	res, err := exec.Command("ssh-keyscan", "-H", hostname).Output()
	if err != nil {
		return fmt.Errorf("failed to run ssh-keyscan for %s: %v", hostname, err)
	}

	err = AppendToFile(ctx, knownHostsPath, string(res), 0644)
	if err != nil {
		return fmt.Errorf("failed to append to known_hosts file %s: %v", knownHostsPath, err)
	}

	logger.Infof("Appended %s to known_hosts file %s\n", hostname, knownHostsPath)
	return nil
}

func AppendSSHConfig(ctx context.Context, configFilePath, hostname, user, identityFile string) error {
	logger, err := GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	// Check if the private key exists
	if _, err := os.Stat(identityFile); err == nil {
		logger.Infof("Identity file %s exists\n", identityFile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check identity file %s: %v", identityFile, err)
	}

	// Write the SSH configuration if config doesnt exist in file
	cfg := config.GetConfig()
	_, err = LoadSSHConfig(ctx, user, cfg.SSH.TimeoutSeconds)
	if err == nil {
		logger.Infof("SSH config for alias %s already exists in %s, skipping append\n", user, configFilePath)
		return nil // Skip if the config already exists
	}

	logger.Infof("Failed to load SSH config for alias %s: %v, appending SSH config to %s\n", user, err, configFilePath)
	configContent := fmt.Sprintf("Host %s\n\tHostName %s\n\tUser %s\n\tIdentityFile %s\n", user, hostname, user, identityFile)
	if err := AppendToFile(ctx, configFilePath, configContent, 0644); err != nil {
		return fmt.Errorf("failed to append SSH config for alias %s: %v", user, err)
	}

	logger.Infof("SSH config for alias %s written to %s\n", user, configFilePath)
	return nil
}
