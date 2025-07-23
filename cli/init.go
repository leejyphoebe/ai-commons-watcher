package cli

import (
	"ai-commons/config"
	"ai-commons/nscc"
	"ai-commons/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI Commons configuration",
	Long:  `This command initializes the AI Commons configuration by creating necessary directories and files.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   initApp,
}

func initApp(cmd *cobra.Command, args []string) {
	fmt.Println("Initializing AI Commons configuration...")

	// check if config flag is set, if it is, load the config file and init
	configFlag := getConfigFlag(cmd, true)

	if configFlag != "" {
		if err := config.InitConfig(configFlag); err != nil {
			fmt.Printf("Error initializing configuration: %v\n", err)
			return
		}
		fmt.Printf("Configuration loaded from %s\n", configFlag)
	} else {
		// Define the configuration directory
		configDir := filepath.Join(os.Getenv("HOME"), ".ai-commons")

		// Create the configuration directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Error creating configuration directory: %v\n", err)
			return
		}

		fmt.Printf("AI Commons configuration initialized at %s\n", configDir)

		// Create a sample configuration file
		configFile := filepath.Join(configDir, "config.yaml")
		sampleConfig := `config_dir: "$HOME/.ai-commons"
node_state_file: "$HOME/.ai-commons/node_state.yaml"
logging:
  level: INFO
  file: ""
  json: false
  stdout: true
ssh:
  hostname: "aspire2antu.nscc.sg"
  config_path: "$HOME/.ai-commons/ssh_config"
  keys_path: "$HOME/.ssh"
  known_hosts_path: "$HOME/.ai-commons/known_hosts"
  key_prefix: "nscc_"
  max_attempts: 3
  sleep_seconds: 3
  timeout_seconds: 30
bitwarden:
  api_url: "https://api.bitwarden.eu"
  identity_url: "https://identity.bitwarden.eu"
  state_file: "$HOME/.ai-commons/bitwarden_state"
`
		if err := os.WriteFile(configFile, []byte(sampleConfig), 0644); err != nil {
			fmt.Printf("Error creating configuration file: %v\n", err)
			return
		}
		fmt.Printf("Sample configuration file created at %s\n", configFile)
		// Initialize the configuration
		if err := config.InitConfig(configFile); err != nil {
			fmt.Printf("Error initializing configuration: %v\n", err)
			return
		}
		fmt.Println("Configuration initialized successfully.")
	}

	// init logger
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}
	logger := utils.GetBaseLogger().WithField("component", "init")

	// create empty bitwarden state file
	cfg := config.GetConfig()
	configDir := cfg.ConfigDir
	bitwardenStateFile := filepath.Join(configDir, "bitwarden_state")
	if err := os.WriteFile(bitwardenStateFile, []byte(""), 0644); err != nil {
		logger.Errorf("Error creating Bitwarden state file: %v", err)
		return
	}
	logger.Infof("Empty Bitwarden state file created at %s", bitwardenStateFile)

	// init ssh keys
	appendKnownHosts := true
	writeSSHConfig := true
	ctx := context.WithValue(context.Background(), utils.LoggerContextKey, logger)
	sshKeys, err := utils.InitSSHKeys(ctx, cfg.SSH.Hostname, appendKnownHosts, writeSSHConfig)
	if err != nil {
		logger.Error("Failed to initialize SSH keys: ", err)
		panic(err)
	}
	logger.Infof("Successfully initialized %d SSH keys", len(sshKeys))

	// record nodes states
	sshConns := make(map[string]*ssh.Client)
	logger.Info("Connecting to SSH hosts...")
	ctx = context.WithValue(ctx, utils.LoggerContextKey, logger)
	states := nscc.NodeStates{}
	states.Nodes = map[string]nscc.NodeState{}
	failedHosts := make([]string, 0)

	for host := range sshKeys {
		conn, err := utils.GetConnection(ctx, host)
		if err != nil {
			logger.Errorf("Failed to connect to host %s: %v", host, err)
			failedHosts = append(failedHosts, host)
			state := nscc.NodeState{
				CanConnect:  false,
				IsReachable: false,
				IsGitSetup:  false,
				Projects:    nil,
				Jobs:        nil,
			}
			states.Nodes[host] = state
			continue
		}
		defer conn.Close()
		sshConns[host] = conn
		logger.Infof("Successfully connected to host %s", host)
		node := nscc.Node{Host: host, Conn: conn}
		state, err := node.GetNodeState(ctx)
		if err != nil {
			logger.Errorf("Failed to get node state for host %s: %v", host, err)
			states.Nodes[host] = state
			continue
		}
		logger.Infof("Node state for host %s: %+v", host, node)
		states.Nodes[host] = state
	}

	logger.Warningf("Failed to connect to the following hosts: %v", failedHosts)

	yamlData, err := yaml.Marshal(states)
	if err != nil {
		logger.Errorf("Failed to marshal node state to YAML: %v", err)
		return
	}

	err = utils.WriteToFile(ctx, cfg.NodeStateFilePath, string(yamlData), 0644)
	if err != nil {
		logger.Errorf("Failed to write node state to file %s: %v", cfg.NodeStateFilePath, err)
		return
	}
	logger.Infof("Node state written to %s", cfg.NodeStateFilePath)
	logger.Debugf("Node state YAML:\n%s", string(yamlData))

	// end
	logger.Info("Run 'ai-commons run' to start using the AI Commons CLI.")
}
