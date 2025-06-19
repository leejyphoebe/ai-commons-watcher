package main

import (
	"ai-commons/config"
	"ai-commons/utils"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
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

	// Initialize logger
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}

	logger := utils.GetBaseLogger().WithField("component", "main")
	logger.Info("Starting ai-commons...")

	// handle signals for graceful shutdown

	// setup environment variables in login node

	// install modules in login node

	// run submit job commands in login node

	// cleanup

	// write tests

}
