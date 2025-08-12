package cli

import (
	"ai-commons/config"
	"ai-commons/general"
	"ai-commons/nscc"
	"ai-commons/utils"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run jobs from config file",
	Long:  `This command runs jobs from configuration file`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runExperiments,
}

func runExperiments(cmd *cobra.Command, args []string) {
	// get config flag
	configFlag := getConfigFlag(cmd, true)
	if configFlag == "" {
		defaultConfigFlag, err := getDefaultConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get default config file: %v\n", err)
		}
		configFlag = defaultConfigFlag
	}
	// load configuration
	if err := config.InitConfig(configFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// init logger
	if err := utils.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	logger := utils.GetBaseLogger().WithField("component", "run")

	logger.Infof("Running experiments from '%s' config file\n", configFlag)

	ctx := context.WithValue(context.Background(), utils.LoggerContextKey, logger)
	cfg := config.GetConfig()

	cluster := utils.GetClusterFromHostname(cfg.SSH.Hostname)

	switch cluster {
	case utils.Aspire2A:
		logger.Infof("Running on NSCC hostname: %s", nscc.NSCCHostname)
	default:
		logger.Infof("Running on custom hostname: %s", cfg.SSH.Hostname)
	}

	err := general.RunJobs(ctx)
	if err != nil {
		logger.Errorf("Failed to run jobs: %v", err)
		os.Exit(1)
	}
	logger.Info("Jobs executed successfully.")
}
