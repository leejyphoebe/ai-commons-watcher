package nscc

import (
	"ai-commons/types"
	"ai-commons/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func GetJobState(ctx context.Context, conn *ssh.Client, host string) ([]types.JobState, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger from context: %w", err)
	}

	// Check if the connection is established
	if conn == nil {
		return nil, fmt.Errorf("SSH connection is nil")
	}

	cmd := "qstat -a"
	out, _, err := utils.RunCommandGetOutput(ctx, cmd, conn)
	if err != nil {
		logger.Errorf("Failed to run command on node %s: %v", host, err)
		return nil, err
	}

	if strings.TrimSpace(out) == "" {
		logger.Warnf("No jobs found for user on node %s", host)
		return nil, nil
	}

	jobs, err := parseQstatOutput(out)
	if err != nil {
		logger.Errorf("Failed to parse qstat output on node %s: %v", host, err)
		return nil, err
	}

	logger.Infof("Successfully fetched %d jobs from qstat on node %s", len(jobs), host)

	return jobs, nil
}

// example qstat -a output
//
// pbs101:
//
// Job ID         Username  Queue    Jobname  SessID  NDS   TSK   Memory   Time  S   Elap Time
// -------------- -------- -------- --------- ------ --- --- -------- -------- - ---------
// 12345.pbs101   user1    default  my_job   12345  1   8   1gb      01:00:00 R   00:05:30
// 12346.pbs101   user2    debug    another_job 0      1   1   2gb      00:30:00 Q   --
// 12347.pbs101   user3    highmem  long_run  67890  2   16  32gb     24:00:00 R   12:45:10
func parseQstatOutput(output string) ([]types.JobState, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no jobs found in qstat output")
	}

	var jobs []types.JobState
	for _, line := range lines[3:] { // Skip header line
		fields := strings.Fields(line)
		if len(fields) != 11 || strings.Contains(fields[0], "---") {
			continue // Skip invalid lines
		}
		job := types.JobState{
			JobID:     strings.TrimSpace(fields[0]),
			JobName:   strings.TrimSpace(fields[3]),
			Queue:     strings.TrimSpace(fields[2]),
			Status:    strings.TrimSpace(fields[9]),
			WallTime:  strings.TrimSpace(fields[8]),
			ElapTime:  strings.TrimSpace(fields[10]),
			Timestamp: time.Now().Format(time.RFC3339),
		}
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		return nil, fmt.Errorf("no valid job entries found in qstat output")
	}

	return jobs, nil
}
