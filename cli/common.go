package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func getConfigFlag(cmd *cobra.Command, optional bool) string {
	configFlag, err := cmd.Flags().GetString("config")
	if !optional && err != nil {
		fmt.Printf("Failed to get config flag: %v\n", err)
	}
	return configFlag
}

func getDefaultConfigPath() (string, error) {
	defaultConfigDir := filepath.Join(os.Getenv("HOME"), ".ai-commons")
	_, err := os.Stat(defaultConfigDir)
	if err != nil {
		fmt.Printf("Default config dir not found in %v: %v.\n", defaultConfigDir, err)
		return "", err
	}
	defaultConfigPath := filepath.Join(defaultConfigDir, "config.yaml")
	_, err = os.Stat(defaultConfigPath)
	if err != nil {
		fmt.Printf("Default config file not found in %v: %v.\n", defaultConfigPath, err)
		return "", err
	}
	return defaultConfigPath, nil
}
