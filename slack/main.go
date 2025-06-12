package main

import (
	"ai-commons/config"
	"ai-commons/utils"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"golang.org/x/crypto/ssh"
)

func main() {

	godotenv.Load(".env")
	slackBotToken := os.Getenv("SLACK_AUTH_TOKEN")
	channelId := os.Getenv("SLACK_CHANNEL_ID")

	// Parse command line flags
	var (
		configFilePath string
		reportTitle    string
	)

	flag.StringVar(&configFilePath, "config", "config.yaml", "Path to the configuration file")
	flag.StringVar(&reportTitle, "title", "Daily Report", "Title for slack report")
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

	// // init SSH keys
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
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	// records the name of the latest file
	latestFilePath := fmt.Sprintf("%s/output/latest_myprojects.txt", cfg.CacheDir)
	// file with the latest data
	myprojectsOutputFilepath := fmt.Sprintf("%s/output/myprojects_%s.csv", cfg.CacheDir, timestamp)

	for host := range sshKeys {
		conn, err := utils.GetConnection(ctx, host)
		if err != nil {
			logger.Errorf("Failed to connect to host %s: %v", host, err)
			panic(err)
		}
		defer conn.Close()
		sshConns[host] = conn
		logger.Infof("Successfully connected to host %s", host)

		// write myprojects_<timestamp>.csv to ./.cache/output/latest_myprojects.txt
		projects, err := utils.GetDailyReportForUser(ctx, conn)
		if err != nil {
			logger.Errorf("Failed to get user daily report: %v", err)
			panic(err)
		}
		utils.AppendMyProjectsToFile(ctx, projects, myprojectsOutputFilepath)
	}
	// read ./.cache/output/latest_myprojects.txt
	prevReportPath, err := utils.ReadFile(ctx, latestFilePath)
	if err != nil {
		logger.Errorf("Failed to read file that stores the path to latest file: %v", err)
	}

	// get daily report string
	message, err := utils.GetDailyReportString(ctx, reportTitle, myprojectsOutputFilepath, prevReportPath)
	if err != nil {
		logger.Errorf("Failed to get daily report string: %v", err)
		panic(err)
	}

	// write to ./.cache/output/myprojects_<timestamp>.csv
	if err := utils.WriteToFile(ctx, latestFilePath, myprojectsOutputFilepath, 0644); err != nil {
		logger.Errorf("Failed to write file: %v", err)
	}

	logger.Info(message)

	// send message to Slack
	api := slack.New(slackBotToken, slack.OptionDebug(true))
	_, _, err = api.PostMessage(channelId, slack.MsgOptionText(message, false))
	if err != nil {
		logger.Errorf("Failed to send message to Slack: %v", err)
	}
}
