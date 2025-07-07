package nscc

import (
	"ai-commons/utils"
	"context"
	"fmt"
)

func (node *Node) GetNodeState(ctx context.Context) (NodeState, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return NodeState{}, fmt.Errorf("failed to get logger from context: %w", err)
	}

	state := NodeState{}

	// Check if the connection is established
	if node.Conn == nil {
		state.CanConnect = false
		logger.Warnf("SSH connection to %s is nil", node.Host)
		return state, nil
	}
	state.CanConnect = true

	// Check if the node is reachable
	cmd := "echo 'Node is reachable'"
	out, _, err := utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		logger.Errorf("Failed to run command on node %s: %v", node.Host, err)
		state.IsReachable = false
		return state, nil
	}
	if out == "" {
		logger.Warnf("Node %s is reachable but returned empty output", node.Host)
		state.IsReachable = false
		return state, nil
	}
	logger.Infof("Node %s is reachable: %s", node.Host, out)
	state.IsReachable = true

	// Check if git is set up
	isGitSetup, err := node.IsGitSetup(ctx)
	if err != nil {
		logger.Error("Failed to check if git is set up: ", err)
		return state, err
	}
	state.IsGitSetup = isGitSetup
	if !isGitSetup {
		logger.Warnf("Git is not set up correctly on node %s", node.Host)
	}

	// get project state
	projects, err := node.GetProjects(ctx)
	if err != nil {
		logger.Errorf("Failed to get projects on node %s: %v", node.Host, err)
		return state, err
	}
	projectStates := make([]ProjectState, 0, len(projects))
	for _, project := range projects {
		if project.CreditSummary.Balance < 0 {
			logger.Warnf("Project %s on node %s has negative credits remaining: %.3f", project.Name, node.Host, project.CreditSummary.Balance)
		}
		projectStates = append(projectStates, ProjectState{
			ProjectName:      project.Name,
			RemainingCredits: project.CreditSummary.Balance,
		})
	}
	state.Projects = projectStates

	// get job state
	jobs, err := node.GetJobStates(ctx)
	if err != nil {
		logger.Errorf("Failed to get job state on node %s: %v", node.Host, err)
		return state, err
	}
	if jobs == nil {
		logger.Warnf("No jobs found for user on node %s", node.Host)
		return state, nil
	}

	logger.Infof("Successfully fetched %d jobs from qstat on node %s", len(jobs), node.Host)
	jobStates := make([]JobState, 0, len(jobs))
	for _, job := range jobs {
		jobStates = append(jobStates, JobState{
			JobID:     job.JobID,
			JobName:   job.JobName,
			Queue:     job.Queue,
			Status:    job.Status,
			WallTime:  job.WallTime,
			ElapTime:  job.ElapTime,
			Timestamp: job.Timestamp,
		})
	}
	state.Jobs = jobStates

	return state, nil
}
