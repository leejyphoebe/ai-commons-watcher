package types

type NodeStates struct {
	Nodes map[string]NodeState `yaml:"nodes"`
}

type NodeState struct {
	CanConnect  bool           `yaml:"can_connect"`
	IsReachable bool           `yaml:"is_reachable"`
	IsGitSetup  bool           `yaml:"is_git_setup"`
	Projects    []ProjectState `yaml:"projects"`
	Jobs        []JobState     `yaml:"jobs"`
}

// NSCC specific
type ProjectState struct {
	ProjectName      string  `yaml:"project_name"`
	RemainingCredits float64 `yaml:"remaining_credits"`
}

type JobState struct {
	JobID     string `yaml:"job_id"`
	JobName   string `yaml:"job_name"`
	Queue     string `yaml:"queue"`
	Status    string `yaml:"status"`
	WallTime  string `yaml:"wall_time"`
	ElapTime  string `yaml:"elap_time"`
	Timestamp string `yaml:"timestamp"`
}
