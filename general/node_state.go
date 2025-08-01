package general

import (
	"ai-commons/nscc"
	"ai-commons/types"
	"ai-commons/utils"
	"context"
	"fmt"
)

func (node *Node) GetNodeState(ctx context.Context) (types.NodeState, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return types.NodeState{}, fmt.Errorf("failed to get logger from context: %w", err)
	}

	state := types.NodeState{}

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
	state.IsGitSetup = isGitSetup
	if !isGitSetup {
		logger.Warnf("Git is not set up correctly on node %s", node.Host)
		if err != nil {
			logger.Warnf("Failed to check if git is set up: %v", err)
		}
	}

	node.Cluster = utils.GetClusterFromHostname(node.Host)
	if node.Cluster == "aspire2a" || node.Cluster == "aspire2p" {
		// get project state
		projects, err := nscc.GetProject(ctx, node.Conn)
		if err != nil {
			logger.Warnf("Failed to get projects on node %s: %v", node.Host, err)
			return state, err
		}
		projectStates := make([]types.ProjectState, 0, len(projects))
		for _, project := range projects {
			if project.CreditSummary.Balance < 0 {
				logger.Warnf("Project %s on node %s has negative credits remaining: %.3f", project.Name, node.Host, project.CreditSummary.Balance)
			}
			projectStates = append(projectStates, types.ProjectState{
				ProjectName:      project.Name,
				RemainingCredits: project.CreditSummary.Balance,
			})
		}
		state.Projects = projectStates

		// get job state
		jobs, err := nscc.GetJobState(ctx, node.Conn, node.Host)
		if err != nil {
			logger.Errorf("Failed to get job state on node %s: %v", node.Host, err)
			return state, err
		}
		if jobs == nil {
			logger.Warnf("No jobs found for user on node %s", node.Host)
			return state, nil
		}

		logger.Infof("Successfully fetched %d jobs from qstat on node %s", len(jobs), node.Host)
		jobStates := make([]types.JobState, 0, len(jobs))
		for _, job := range jobs {
			jobStates = append(jobStates, types.JobState{
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
	}

	return state, nil
}
