package nscc

import (
	"ai-commons/utils"
	"context"
	"fmt"
	"strings"
	"time"
)

func (node *Node) GetJobStates(ctx context.Context) ([]JobState, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger from context: %w", err)
	}

	// Check if the connection is established
	if node.Conn == nil {
		return nil, fmt.Errorf("SSH connection is nil")
	}

	cmd := "qstat -a"
	out, _, err := utils.RunCommandGetOutput(ctx, cmd, node.Conn)
	if err != nil {
		logger.Errorf("Failed to run command on node %s: %v", node.Host, err)
		return nil, err
	}

	if strings.TrimSpace(out) == "" {
		logger.Warnf("No jobs found for user on node %s", node.Host)
		return nil, nil
	}

	jobs, err := parseQstatOutput(out)
	if err != nil {
		logger.Errorf("Failed to parse qstat output on node %s: %v", node.Host, err)
		return nil, err
	}

	logger.Infof("Successfully fetched %d jobs from qstat on node %s", len(jobs), node.Host)

	return jobs, nil
}

// Job ID         Username  Queue    Jobname  SessID  NDS   TSK   Memory   Time  S   Elap Time
// -------------- -------- -------- --------- ------ --- --- -------- -------- - ---------
// 12345.server.com user1    default  my_job   12345  1   8   1gb      01:00:00 R   00:05:30
// 12346.server.com user2    debug    another_job 0      1   1   2gb      00:30:00 Q   --
// 12347.server.com user3    highmem  long_run  67890  2   16  32gb     24:00:00 R   12:45:10
func parseQstatOutput(output string) ([]JobState, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no jobs found in qstat output")
	}

	var jobs []JobState
	for _, line := range lines[1:] { // Skip header line
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue // Skip invalid lines
		}
		job := JobState{
			JobID:     fields[0],
			JobName:   fields[3],
			Queue:     fields[2],
			Status:    fields[9],
			WallTime:  fields[8],
			ElapTime:  fields[10],
			Timestamp: time.Now().Format(time.RFC3339),
		}
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		return nil, fmt.Errorf("no valid job entries found in qstat output")
	}

	return jobs, nil
}
