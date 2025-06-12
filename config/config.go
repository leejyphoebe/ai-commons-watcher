package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type LoggingConfig struct {
	File   string `yaml:"file"`
	Level  string `yaml:"level"`
	Json   bool   `yaml:"json"`
	Stdout bool   `yaml:"stdout"`
}

type SSHInitConfig struct {
	MasterHost     string `yaml:"master_host"` // required, enter your username on nscc
	Hostname       string `yaml:"hostname"`
	ConfigPath     string `yaml:"config_path"`
	KeysPath       string `yaml:"keys_path"`
	KnownHostsPath string `yaml:"known_hosts_path"`
}

type BitwardenConfig struct {
	ApiUrl      string `yaml:"api_url"`
	IdentityUrl string `yaml:"identity_url"`
	StateFile   string `yaml:"state_file"`
}

type ModelConfig struct {
	Name    string `yaml:"name"`    // name of the model
	Version string `yaml:"version"` // version of the model
	Source  string `yaml:"source"`  // source of the model, e.g., "huggingface", "local", etc.
	Dir     string `yaml:"dir"`     // directory where the model is stored
}

type EnvConfig struct {
	Key   string `yaml:"key"`   // environment variable key
	Value string `yaml:"value"` // environment variable value
}

type NsccConfig struct {
	CopyDir         string      `yaml:"copy_dir"`         // directory to copy files to the NSCC host
	SifImagesDir    string      `yaml:"sif_images_dir"`   // directory to store Singularity SIF images
	ResultDir       string      `yaml:"result_dir"`       // directory to store results from NSCC
	Model           ModelConfig `yaml:"model"`            // model configuration for NSCC
	RequiredModules []string    `yaml:"required_modules"` // list of required modules to be installed on NSCC
	SetupSteps      []string    `yaml:"setup_steps"`      // list of setup steps to be executed on NSCC
	SetupScript     string      `yaml:"setup_script"`     // script to be executed on NSCC for setup
	Envs            []EnvConfig `yaml:"envs"`             // environment variables to be set on NSCC
}

type Config struct {
	CacheDir               string          `yaml:"cache_dir"`
	NsccUsageCacheFilePath string          `yaml:"nscc_usage_cache_file_path"`
	SSH                    SSHInitConfig   `yaml:"ssh"`
	Logging                LoggingConfig   `yaml:"logging"`
	Bitwarden              BitwardenConfig `yaml:"bitwarden"`
	NSCC                   NsccConfig      `yaml:"nscc"`
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
	if cfg.CacheDir == "" {
		cfg.CacheDir = "$PWD/.cache"
	}
	if cfg.NsccUsageCacheFilePath == "" {
		cfg.NsccUsageCacheFilePath = "$PWD/.cache/nscc_usage.csv"
	}
	// if cfg.Nscc.ProjectRepo == "" {
	// 	return fmt.Errorf("project_repo is required in the configuration")
	// }
	if cfg.SSH.Hostname == "" {
		cfg.SSH.Hostname = "aspire2antu.nscc.sg"
	}
	if cfg.SSH.MasterHost == "" {
		return fmt.Errorf("ssh.master_host is required in the configuration")
	}
	if cfg.SSH.ConfigPath == "" {
		cfg.SSH.ConfigPath = "$PWD/.cache/ssh_config"
	}
	if cfg.SSH.KeysPath == "" {
		cfg.SSH.KeysPath = "$HOME/.ssh"
	}
	if cfg.SSH.KnownHostsPath == "" {
		cfg.SSH.KnownHostsPath = "$PWD/.cache/known_hosts"
	}
	if cfg.Bitwarden.ApiUrl == "" {
		cfg.Bitwarden.ApiUrl = "https://api.bitwarden.eu"
	}
	if cfg.Bitwarden.IdentityUrl == "" {
		cfg.Bitwarden.IdentityUrl = "https://identity.bitwarden.eu"
	}
	if cfg.Bitwarden.StateFile == "" {
		cfg.Bitwarden.StateFile = "$PWD/.cache/state"
	}
	if len(cfg.NSCC.RequiredModules) == 0 {
		cfg.NSCC.RequiredModules = []string{"git", "ssh", "singularity"}
	}

	// Expand environment variables in paths
	cfg.CacheDir = os.ExpandEnv(cfg.CacheDir)
	cfg.SSH.ConfigPath = os.ExpandEnv(cfg.SSH.ConfigPath)
	cfg.SSH.KeysPath = os.ExpandEnv(cfg.SSH.KeysPath)
	cfg.SSH.KnownHostsPath = os.ExpandEnv(cfg.SSH.KnownHostsPath)
	cfg.Bitwarden.StateFile = os.ExpandEnv(cfg.Bitwarden.StateFile)

	// Ensure the cache directory exists
	if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s: %w", cfg.CacheDir, err)
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
