package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type LoggingConfig struct {
	File   string `yaml:"file"`
	Level  string `yaml:"level"`
	Json   *bool  `yaml:"json"`
	Stdout *bool  `yaml:"stdout"`
}

type SSHInitConfig struct {
	Hostname       string `yaml:"hostname"`
	ConfigPath     string `yaml:"config_path"`
	KeysPath       string `yaml:"keys_path"`
	KnownHostsPath string `yaml:"known_hosts_path"`
	KeyPrefix      string `yaml:"key_prefix"`      // prefix for SSH keys on bitwarden and downloaded keys, e.g., "nscc_"
	MaxAttempts    *int   `yaml:"max_attempts"`    // maximum attempts to connect to SSH host
	SleepSeconds   *int   `yaml:"sleep_seconds"`   // seconds to sleep between connection attempts
	TimeoutSeconds *int   `yaml:"timeout_seconds"` // timeout for SSH connection in seconds
}

type BitwardenConfig struct {
	ApiUrl      string `yaml:"api_url"`
	IdentityUrl string `yaml:"identity_url"`
	StateFile   string `yaml:"state_file"`
}

type ExperimentConfigPath struct {
	Src     string `yaml:"src"`     // source directory for the experiment
	Dest    string `yaml:"dest"`    // destination directory for the experiment on NSCC
	Command string `yaml:"command"` // command to run the experiment
}

type ExperimentNodeConfig struct {
	Host                  string                 `yaml:"host"`                    // hostname of the node
	ExperimentConfigPaths []ExperimentConfigPath `yaml:"experiment_config_paths"` // list of experiment configurations for the node
}

type SetupFile struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

type ExperimentSetup struct {
	SetupFilesDir         string      `yaml:"setup_files_dir"`
	SetupScriptLocalPath  string      `yaml:"setup_script_local_path"`
	SetupScriptRemotePath string      `yaml:"setup_script_remote_path"`
	SetupFiles            []SetupFile `yaml:"setup_files"`
	SetupArgs             []string    `yaml:"setup_args"` // arguments for the setup script
}
type ExperimentCleanup struct {
	CleanupScript           string `yaml:"cleanup_script"`
	CleanupScriptRemotePath string `yaml:"cleanup_script_remote_path"`
}
type ExperimentsConfig struct {
	Name              string                 `yaml:"name"`              // name of the job
	Nodes             []ExperimentNodeConfig `yaml:"nodes"`             // list of nodes to run the job on
	LocalConfigDir    string                 `yaml:"local_config_dir"`  // local directory for job configuration
	RemoteConfigDir   string                 `yaml:"remote_config_dir"` // remote directory for job configuration on NSCC
	CmdDir            string                 `yaml:"cmd_dir"`           // directory where the job command is located
	ExperimentSetup   ExperimentSetup        `yaml:"setup"`             // configuration for the experiment setup
	ExperimentCleanup ExperimentCleanup      `yaml:"cleanup"`
	GitRequired       bool                   `yaml:"git_required"` // whether git is required for the job
}

type Config struct {
	ConfigDir         string              `yaml:"config_dir"`
	NodeStateFilePath string              `yaml:"node_state_file"`
	SSH               SSHInitConfig       `yaml:"ssh"`
	Logging           LoggingConfig       `yaml:"logging"`
	Bitwarden         BitwardenConfig     `yaml:"bitwarden"`
	Experiments       []ExperimentsConfig `yaml:"experiments"`
}

var config *Config

// LoadConfig reads the YAML configuration from the specified file path.
func LoadConfigFromFile(filePath string) (*Config, error) {
	// Read the YAML file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Initialize an empty AppConfig struct
	var cfg Config

	// Unmarshal the YAML data into the struct
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return &cfg, nil
}

func InitConfig(filePath string) error {
	if config != nil {
		return fmt.Errorf("configuration already initialized")
	}

	cfg, err := LoadConfigFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate required fields and set defaults
	if cfg.ConfigDir == "" {
		cfg.ConfigDir = "$HOME/.ai-commons"
	}
	if cfg.NodeStateFilePath == "" {
		cfg.NodeStateFilePath = "$HOME/.ai-commons/node_state.yaml"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "INFO"
	}
	if cfg.Logging.File == "" && cfg.Logging.Stdout == nil {
		cfg.Logging.Stdout = new(bool)
		*cfg.Logging.Stdout = true // default to stdout if no file is specified
	}
	if cfg.Logging.Stdout == nil {
		cfg.Logging.Stdout = new(bool)
		*cfg.Logging.Stdout = true // default to true if not specified
	}
	if cfg.Logging.Json == nil {
		cfg.Logging.Json = new(bool)
		*cfg.Logging.Json = false // default to false if not specified
	}
	if cfg.SSH.Hostname == "" {
		cfg.SSH.Hostname = "aspire2antu.nscc.sg"
	}
	if cfg.SSH.ConfigPath == "" {
		cfg.SSH.ConfigPath = "$HOME/.ai-commons/ssh_config"
	}
	if cfg.SSH.KeysPath == "" {
		cfg.SSH.KeysPath = "$HOME/.ssh"
	}
	if cfg.SSH.KnownHostsPath == "" {
		cfg.SSH.KnownHostsPath = "$HOME/.ai-commons/known_hosts"
	}
	if cfg.SSH.KeyPrefix == "" {
		cfg.SSH.KeyPrefix = "nscc_"
	}
	if cfg.SSH.MaxAttempts == nil {
		cfg.SSH.MaxAttempts = new(int)
		*cfg.SSH.MaxAttempts = 1 // default to 1 if not specified
	}
	if cfg.SSH.SleepSeconds == nil {
		cfg.SSH.SleepSeconds = new(int)
		*cfg.SSH.SleepSeconds = 0 // default to 0 seconds if not specified
	}
	if cfg.SSH.TimeoutSeconds == nil {
		cfg.SSH.TimeoutSeconds = new(int)
		*cfg.SSH.TimeoutSeconds = 30 // default to 30 seconds if not specified
	}
	if cfg.Bitwarden.ApiUrl == "" {
		cfg.Bitwarden.ApiUrl = "https://api.bitwarden.eu"
	}
	if cfg.Bitwarden.IdentityUrl == "" {
		cfg.Bitwarden.IdentityUrl = "https://identity.bitwarden.eu"
	}
	if cfg.Bitwarden.StateFile == "" {
		cfg.Bitwarden.StateFile = "$HOME/.ai-commons/bitwarden_state"
	}

	// validate experiment config
	for i, exp := range cfg.Experiments {
		if exp.Name == "" {
			return fmt.Errorf("experiments[%d].name is required in the configuration", i)
		}
		if exp.GitRequired && len(exp.Nodes) == 0 {
			return fmt.Errorf("experiments[%d] requires git but no nodes are specified", i)
		}
		if len(exp.Nodes) == 0 {
			return fmt.Errorf("experiments[%d].nodes is required in the configuration", i)
		}
		if exp.LocalConfigDir == "" {
			return fmt.Errorf("experiments[%d].local_config_dir is required in the configuration", i)
		}
		if exp.RemoteConfigDir == "" {
			return fmt.Errorf("experiments[%d].remote_config_dir is required in the configuration", i)
		}
		if exp.CmdDir == "" {
			return fmt.Errorf("experiments[%d].cmd_dir is required in the configuration", i)
		}
		if exp.ExperimentSetup.SetupFilesDir == "" {
			exp.ExperimentSetup.SetupFilesDir = "/scratch/users/ntu/$USER/.ai-commons/setup"
		}
		if exp.ExperimentSetup.SetupScriptLocalPath == "" {
			return fmt.Errorf("experiments[%d].setup.setup_script_local_path is required in the configuration", i)
		}
		if exp.ExperimentSetup.SetupScriptRemotePath == "" {
			return fmt.Errorf("experiments[%d].setup.setup_script_remote_path is required in the configuration", i)
		}
		// Validate each experiment config
		for j, exp := range exp.Nodes {
			if exp.Host == "" {
				return fmt.Errorf("experiments[%d].nodes[%d].host is required in the configuration", i, j)
			}
			if len(exp.ExperimentConfigPaths) == 0 {
				return fmt.Errorf("experiments[%d].nodes[%d].experiment_configs is required in the configuration", i, j)
			}
			// Validate each experiment config path
			for k, expPath := range exp.ExperimentConfigPaths {
				if expPath.Src == "" {
					return fmt.Errorf("experiments[%d].nodes[%d].experiment_configs[%d].src is required in the configuration", i, j, k)
				}
				if expPath.Dest == "" {
					return fmt.Errorf("experiments[%d].nodes[%d].experiment_configs[%d].dest is required in the configuration", i, j, k)
				}
				if expPath.Command == "" {
					return fmt.Errorf("experiments[%d].nodes[%d].experiment_configs[%d].command is required in the configuration", i, j, k)
				}
			}
		}
	}

	// Expand environment variables in paths
	cfg.ConfigDir = os.ExpandEnv(cfg.ConfigDir)
	cfg.SSH.ConfigPath = os.ExpandEnv(cfg.SSH.ConfigPath)
	cfg.SSH.KeysPath = os.ExpandEnv(cfg.SSH.KeysPath)
	cfg.SSH.KnownHostsPath = os.ExpandEnv(cfg.SSH.KnownHostsPath)
	cfg.Bitwarden.StateFile = os.ExpandEnv(cfg.Bitwarden.StateFile)
	cfg.NodeStateFilePath = os.ExpandEnv(cfg.NodeStateFilePath)

	// Ensure the config directory exists
	if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", cfg.ConfigDir, err)
	}

	config = cfg
	return nil
}

func GetConfig() *Config {
	if config == nil {
		panic("application configuration not initialized. Call config.Initialize() first.")
	}
	return config
}
