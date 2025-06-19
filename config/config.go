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

type ExperimentConfig struct {
	Src  string `yaml:"src"`  // source directory for the experiment
	Dest string `yaml:"dest"` // destination directory for the experiment on NSCC
}

type JobConfig struct {
	Name                  string             `yaml:"name"`                // name of the job
	Description           string             `yaml:"description"`         // description of the job
	NumJobsPerNode        int                `yaml:"num_jobs_per_node"`   // number of jobs to run per node
	NumNodes              int                `yaml:"num_nodes"`           // number of nodes to use for the job
	Nodes                 []string           `yaml:"nodes"`               // list of nodes to run the job on
	LocalConfigDir        string             `yaml:"local_config_dir"`    // local directory for job configuration
	RemoteConfigDir       string             `yaml:"remote_config_dir"`   // remote directory for job configuration on NSCC
	ExperimentConfigs     []ExperimentConfig `yaml:"experiment_configs"`  // list of experiment configurations for the job
	CmdDir                string             `yaml:"cmd_dir"`             // directory where the job command is located
	Command               string             `yaml:"command"`             // command to run the job
	SetupScriptPath       string             `yaml:"setup_script"`        // path to the setup script to be executed before running the job
	RemoteSetupScriptPath string             `yaml:"remote_setup_script"` // path to the setup script on the remote node
}

type Config struct {
	CacheDir               string          `yaml:"cache_dir"`
	NsccUsageCacheFilePath string          `yaml:"nscc_usage_cache_file_path"`
	NodeStateFilePath      string          `yaml:"node_state_file_path"`
	SSH                    SSHInitConfig   `yaml:"ssh"`
	Logging                LoggingConfig   `yaml:"logging"`
	Bitwarden              BitwardenConfig `yaml:"bitwarden"`
	NSCC                   NsccConfig      `yaml:"nscc"`
	Jobs                   []JobConfig     `yaml:"jobs"`
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
	if cfg.NodeStateFilePath == "" {
		cfg.NodeStateFilePath = "$PWD/.cache/node_state.yaml"
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

	// validate job config
	for i, job := range cfg.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d].name is required in the configuration", i)
		}
		if job.Description == "" {
			return fmt.Errorf("job[%d].description is required in the configuration", i)
		}
		if job.NumJobsPerNode <= 0 {
			return fmt.Errorf("job[%d].num_jobs_per_node must be greater than 0", i)
		}
		if job.NumNodes <= 0 {
			return fmt.Errorf("job[%d].num_nodes must be greater than 0", i)
		}
		if job.LocalConfigDir == "" {
			return fmt.Errorf("job[%d].local_config_dir is required in the configuration", i)
		}
		if job.RemoteConfigDir == "" {
			return fmt.Errorf("job[%d].remote_config_dir is required in the configuration", i)
		}
		if job.CmdDir == "" {
			return fmt.Errorf("job[%d].cmd_dir is required in the configuration", i)
		}
		if job.Command == "" {
			return fmt.Errorf("job[%d].command is required in the configuration", i)
		}
		if job.SetupScriptPath == "" {
			return fmt.Errorf("job[%d].setup_script is required in the configuration", i)
		}
		if job.RemoteSetupScriptPath == "" {
			return fmt.Errorf("job[%d].remote_setup_script is required in the configuration", i)
		}
		if len(job.ExperimentConfigs) == 0 {
			return fmt.Errorf("job[%d].experiment_configs must contain at least one experiment configuration", i)
		}
		if len(job.ExperimentConfigs) != job.NumJobsPerNode*len(job.Nodes) {
			return fmt.Errorf("job[%d].experiment_configs length must match num_jobs_per_node * len(job.nodes)", i)
		}
		// Validate each experiment config
		for j, exp := range job.ExperimentConfigs {
			if exp.Src == "" {
				return fmt.Errorf("job[%d].experiment_configs[%d].src is required in the configuration", i, j)
			}
			if exp.Dest == "" {
				return fmt.Errorf("job[%d].experiment_configs[%d].dest is required in the configuration", i, j)
			}
		}
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
