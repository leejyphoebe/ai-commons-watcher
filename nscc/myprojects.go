package nscc

import (
	"ai-commons/utils"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Project struct {
	Name          string
	Username      string
	LastUpdated   time.Time
	CreditSummary CreditSummary
	Usage         []CreditUsageByType
}

type CreditSummary struct {
	Unit    string
	Grant   float64
	Used    float64
	Balance float64
	InDoubt float64
}

type CreditUsageByType struct {
	Unit   string
	Usage  float64
	SURate float64
	SUUsed float64
}

type MyProjectsSummary struct {
	Timestamp   time.Time // This should be set based on the current date and time
	LastUpdated time.Time
	Username    string
	ProjectName string
	Grant       float64
	Used        float64
	Balance     float64
	InDoubt     float64
	GPUHour     float64
	CPUHour     float64
}

var tz = "Asia/Singapore"
var timeFormat = "2006-01-02 15:04:05"

func GetProject(ctx context.Context, conn *ssh.Client) ([]Project, error) {
	// Get the logger from the context
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	timezone, err := time.LoadLocation(tz)
	if err != nil {
		logger.Errorf("Failed to load timezone %s: %v", tz, err)
		return nil, fmt.Errorf("failed to load timezone %s: %v", tz, err)
	}
	logger.Infof("Using timezone: %s", timezone)
	timestamp := time.Now().In(timezone).Format(timeFormat)
	logger.Infof("Generating daily report for %s", timestamp)

	// run myprojects to check credits
	cmd := "myprojects"
	stdout, _, err := utils.RunCommandGetOutput(ctx, cmd, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to run credit check command: %v", err)
	}
	if stdout == "" {
		return nil, fmt.Errorf("credit check command returned empty output")
	}
	lines := strings.Split(stdout, "\n")
	lines = append(lines, "---END---\n") // hax to mark the end of the output
	projects, err := parseMyProjectsStdout(ctx, conn.User(), lines)
	if err != nil {
		return nil, fmt.Errorf("failed to parse myprojects output: %v", err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found in myprojects output")
	}

	return projects, nil
}

// example stdout:
// Project : personal-silv0011
// ExpDate : 2032-12-31
// Organis : NTU
// P.Inves : SILVIANA
// Appl.Id : null
// P.Title : Personal Project

// Project personal-silv0011 balance as of 12/06/2025-10:39:23
// +--------+---------------------+---------------------+---------------------+---------------------+
// | Unit   |               Grant |                Used |             Balance |            In Doubt |
// +--------+---------------------+---------------------+---------------------+---------------------+
// | SU     |          100000.000 |             122.346 |           99877.654 |               0.000 |
// +--------+---------------------+---------------------+---------------------+---------------------+
// In doubt - SU deducted for current running jobs (Prepaid)

// Project personal-silv0011 SU Usage breakdown
// +---------------+-----------------+-----------------+-----------------+
// | Unit          |           Usage |         SU Rate |         SU Used |
// +---------------+-----------------+-----------------+-----------------+
// | CPU Hour      |           0.000 |               1 |           0.000 |
// | GPU Hour      |           1.911 |              64 |         122.346 |
// +---------------+-----------------+-----------------+-----------------+

func splitTableLine(line string) []string {
	parts := strings.Split(line, "|")
	var fields []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			fields = append(fields, p)
		}
	}
	return fields
}

func parseMyProjectsStdout(ctx context.Context, username string, lines []string) ([]Project, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	logger.Debugf("Parsing myprojects output: %v", lines)

	var projects []Project
	var currentProject *Project
	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "---END---") {
			if currentProject != nil {
				projects = append(projects, *currentProject)
				currentProject = nil
				currentSection = ""
			}
			continue
		}
		if strings.HasPrefix(line, "Project\t:") {
			currentProject = &Project{
				Name:     strings.TrimSpace(strings.TrimPrefix(line, "Project\t:")),
				Username: username,
			}
			currentSection = ""
			continue
		}

		if strings.Contains(line, "balance as of") {
			if currentProject == nil {
				return nil, fmt.Errorf("found balance line without a project")
			}
			// Extract the last updated time from the line
			parts := strings.Split(line, "balance as of")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid balance line: %s, expected format 'Project <name> balance as of <date>'", line)
			}
			dateStr := strings.TrimSpace(parts[1])
			// format: 12/06/2025-10:39:23
			currentProject.LastUpdated, err = time.Parse("02/01/2006-15:04:05", dateStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse date from balance line: %s, error: %v", dateStr, err)
			}
			logger.Debugf("Parsed project %s last updated time: %s", currentProject.Name, currentProject.LastUpdated)
			if currentProject.LastUpdated.IsZero() {
				return nil, fmt.Errorf("project %s has an invalid last updated time", currentProject.Name)
			}
			continue
		}

		if strings.Contains(line, "Grant") {
			if currentProject == nil {
				return nil, fmt.Errorf("found credit summary without a project")
			}
			currentSection = "CreditSummary"
			continue
		}

		if strings.Contains(line, "SU Used") {
			if currentProject == nil {
				return nil, fmt.Errorf("found SU Usage without a project")
			}
			currentSection = "CreditUsageByType"
			continue
		}

		if strings.HasPrefix(line, "+") {
			continue // Skip table borders
		}

		if strings.HasPrefix(line, "|") {
			switch currentSection {
			case "CreditSummary":
				parts := splitTableLine(line)
				if len(parts) < 5 {
					return nil, fmt.Errorf("invalid credit summary line: %s, expected 5 fields", line)
				}

				values := make([]float64, 4)
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
					if i > 0 {
						values[i-1], _ = strconv.ParseFloat(parts[i], 64)
					}
				}

				currentProject.CreditSummary = CreditSummary{
					Unit:    parts[0],
					Grant:   values[0],
					Used:    values[1],
					Balance: values[2],
					InDoubt: values[3],
				}
			case "CreditUsageByType":
				parts := splitTableLine(line)
				if len(parts) < 4 {
					return nil, fmt.Errorf("invalid SU usage line: %s, expected 4 fields", line)
				}

				values := make([]float64, 3)
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
					if i > 0 {
						values[i-1], _ = strconv.ParseFloat(parts[i], 64)
					}
				}
				currentProject.Usage = append(currentProject.Usage, CreditUsageByType{
					Unit:   parts[0],
					Usage:  values[0],
					SURate: values[1],
					SUUsed: values[2],
				})
			}
		}
	}
	logger.Debugf("Parsed %d projects: %+v", len(projects), projects)
	return projects, nil
}

func createMyProjectSummary(ctx context.Context, project Project) (MyProjectsSummary, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return MyProjectsSummary{}, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	if project.CreditSummary.Unit == "" {
		return MyProjectsSummary{}, fmt.Errorf("project %s has empty credit summary unit", project.Name)
	}
	if len(project.Usage) == 0 {
		return MyProjectsSummary{}, fmt.Errorf("project %s has no usage data", project.Name)
	}

	timezone, err := time.LoadLocation(tz)
	if err != nil {
		logger.Errorf("Failed to load timezone %s: %v", tz, err)
		return MyProjectsSummary{}, fmt.Errorf("failed to load timezone %s: %v", tz, err)
	}
	logger.Infof("Using timezone: %s", timezone)
	timestamp := time.Now().In(timezone)

	output := MyProjectsSummary{
		Username:    project.Username,
		ProjectName: project.Name,
		Grant:       project.CreditSummary.Grant,
		Used:        project.CreditSummary.Used,
		Balance:     project.CreditSummary.Balance,
		InDoubt:     project.CreditSummary.InDoubt,
		LastUpdated: project.LastUpdated,
		Timestamp:   timestamp, // Set the current timestamp
	}

	for _, usage := range project.Usage {
		switch usage.Unit {
		case "GPU Hour":
			output.GPUHour = usage.SUUsed
		case "CPU Hour":
			output.CPUHour = usage.SUUsed
		}
	}

	logger.Debugf("Created MyProjectsSummary: %+v", output)
	return output, nil
}

// save summary and breakdown as csv
func SaveSummaryToCsv(ctx context.Context, projects []Project, filePath string) error {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	logger.Infof("Writing myprojects data to %s", filePath)

	var sb strings.Builder
	isFileExists, err := utils.CheckIfFileExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists %s: %v", filePath, err)
	}

	if !isFileExists {
		sb.WriteString("timestamp,last_updated,username,projectname,grant,used,balance,indoubt,gpu_hour,cpu_hour\n")
	} else {
		// If the file exists, we assume it already has the header
		logger.Infof("File %s already exists, appending data", filePath)
	}

	for _, project := range projects {
		output, err := createMyProjectSummary(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to create MyProjectsSummary for project %s: %v", project.Name, err)
		}
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%.3f,%.3f,%.3f,%.3f,%.3f,%.3f",
			output.Timestamp.Format(timeFormat),
			output.LastUpdated.Format(timeFormat),
			output.Username,
			output.ProjectName,
			output.Grant,
			output.Used,
			output.Balance,
			output.InDoubt,
			output.GPUHour,
			output.CPUHour,
		))
		logger.Debugf("Appended project %s data to string builder", project.Name)
	}

	return utils.AppendToFile(ctx, filePath, sb.String(), 0644)
}

func sortOutputByBalanceDesc(outputs []MyProjectsSummary) []MyProjectsSummary {
	sort.Slice(outputs[:], func(i, j int) bool {
		return outputs[i].Balance > outputs[j].Balance
	})
	return outputs
}

func parseMyProjectsCsv(line string) (MyProjectsSummary, error) {
	if strings.HasPrefix(line, "timestamp") || line == "" {
		// Skip header line
		return MyProjectsSummary{}, nil
	}
	fields := strings.Split(line, ",")
	if len(fields) < 9 {
		return MyProjectsSummary{}, fmt.Errorf("malformed line: %s, expected at least 9 fields", line)
	}

	timestamp, err := time.Parse(timeFormat, fields[0])
	if err != nil {
		return MyProjectsSummary{}, fmt.Errorf("invalid timestamp in line: %s", line)
	}

	lastUpdated, err := time.Parse(timeFormat, fields[1])
	if err != nil {
		return MyProjectsSummary{}, fmt.Errorf("invalid last updated time in line: %s", line)
	}

	values := make([]float64, 6)
	for i := range fields {
		if i < 4 {
			continue
		}
		fields[i] = strings.TrimSpace(fields[i])
		if i > 0 {
			values[i-4], _ = strconv.ParseFloat(fields[i], 64)
		}
	}

	output := MyProjectsSummary{
		Timestamp:   timestamp,
		LastUpdated: lastUpdated,
		Username:    fields[2],
		ProjectName: fields[3],
		Grant:       values[0],
		Used:        values[1],
		Balance:     values[2],
		InDoubt:     values[3],
		GPUHour:     values[4],
		CPUHour:     values[5],
	}

	return output, nil
}

func readOutputFileToList(ctx context.Context, filePath string) ([]MyProjectsSummary, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	logger.Infof("Reading myprojects data from %s", filePath)

	data, err := utils.ReadFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	lines := strings.Split(data, "\n")
	outputs := make([]MyProjectsSummary, 0)

	for _, line := range lines {
		output, err := parseMyProjectsCsv(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line: %s, error: %v", line, err)
		}
		if output.ProjectName != "" {
			outputs = append(outputs, output)
			logger.Debugf("Parsed output for project %s: %+v", output.ProjectName, output)
		}
	}
	// Sort outputs by balance in descending order
	outputs = sortOutputByBalanceDesc(outputs)
	logger.Info("Sorted outputs by balance in descending order")
	logger.Info(outputs)
	if len(outputs) == 0 {
		logger.Warnf("No valid projects found in file %s", filePath)
		return nil, fmt.Errorf("no valid projects found in file %s", filePath)
	}
	logger.Infof("Finished reading %d projects from file %s", len(outputs), filePath)
	return outputs, nil
}

func readOutputFileToMap(ctx context.Context, filePath string) (map[string]MyProjectsSummary, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	logger.Infof("Reading myprojects data from %s", filePath)

	data, err := utils.ReadFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	lines := strings.Split(data, "\n")
	outputs := make(map[string]MyProjectsSummary)

	for _, line := range lines {
		output, err := parseMyProjectsCsv(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line: %s, error: %v", line, err)
		}
		if output.ProjectName != "" {
			outputs[output.ProjectName] = output
			logger.Debugf("Parsed output for project %s: %+v", output.ProjectName, output)
		}
	}
	logger.Infof("Finished reading %d projects from file %s", len(outputs), filePath)
	return outputs, nil
}

// Format:
// 📊 *<title>*
// 🕑 _Fetched on: <datetime>_
//
//  1. <username> — 🪙 <balance> (🕺/🔻 <yesterday's_balance>)
//  2. <username> — 🪙 <balance> (🕺/🔻 <yesterday's_balance>)
//     ...
func GetDailyReportString(ctx context.Context, title string, newFilePath, prevFilePath string, failedHosts []string) (string, error) {
	logger, err := utils.GetLoggerFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve logger from context: %v", err)
	}

	logger.Infof("Generating daily report string from %s", newFilePath)

	new, err := readOutputFileToList(ctx, newFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read today's output file %s: %v", newFilePath, err)
	}

	if prevFilePath == "" {
		logger.Warn("Yesterday's output file path is empty, skipping previous data")
		prevFilePath = newFilePath // Use new file as a fallback
	}
	prev, err := readOutputFileToMap(ctx, prevFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read previous output file %s: %v", prevFilePath, err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *%s*\n", strings.TrimSpace(title)))
	loc, err := time.LoadLocation(tz)
	if err != nil {
		logger.Errorf("Error loading location: %v", err)
		return "", fmt.Errorf("failed to load location: %v", err)
	}
	sb.WriteString(fmt.Sprintf("🕑 _Fetched on: %s_\n", time.Now().In(loc).Format(timeFormat)))
	sb.WriteString("\n")

	// summarize projects with full credits at the end
	untouchedProjects := make(map[string]MyProjectsSummary)
	for _, output := range new {
		if output.Balance == output.Grant {
			untouchedProjects[output.ProjectName] = output
		}
	}

	if len(new) > 0 {
		sb.WriteString("💸 *Accounts with Credit Usage:*\n")
	}

	under1k := false

	i := 1
	prevTimestamp := time.Time{}
	for _, newOutput := range new {
		if newOutput.Balance == 0 || newOutput.Used == 0 {
			continue // Skip projects with no balance or usage
		}
		prevOutput, exists := prev[newOutput.ProjectName]
		if !exists {
			prevOutput = MyProjectsSummary{}
		}
		if prevTimestamp.IsZero() {
			if !prevOutput.Timestamp.IsZero() {
				prevTimestamp = prevOutput.Timestamp
			} else {
				prevTimestamp = time.Time{}
			}
		}

		if newOutput.Balance < 1000 && !under1k {
			under1k = true
			sb.WriteString("\n")
			sb.WriteString("🥀 *Accounts with Low Balance (< 1k):*\n")
		}

		balanceChange := prevOutput.Balance - newOutput.Balance
		sb.WriteString(fmt.Sprintf("%d. *%s*", i, newOutput.Username))
		sb.WriteString(" — 🪙 ")

		sb.WriteString(utils.FormatToK(newOutput.Balance, 1, "%.1fk", "%.0f"))
		if balanceChange != 0 {
			if balanceChange > 0 {
				sb.WriteString(" (🔻 ")
			} else {
				sb.WriteString(" (:gopher_dance: ")
			}
			sb.WriteString(utils.FormatToK(math.Abs(balanceChange), 1, "%.1fk)", "%.0f)"))
		}
		sb.WriteString("\n")
		i++
	}

	if len(untouchedProjects) > 0 {
		sb.WriteString("\n")
		sb.WriteString("🥛 *Full Credit Account")
		if len(untouchedProjects) > 1 {
			sb.WriteString("s")
		}
		sb.WriteString("*:\n")
		for _, untouched := range untouchedProjects {
			sb.WriteString(fmt.Sprintf("%d. *%s*\n", i, untouched.Username))
			i++
		}
	}

	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("💼 Total Projects: %d\n", len(new)))

	if len(new) == 0 {
		sb.WriteString("No projects found for this run.\n")
	}
	if len(prev) == 0 {
		sb.WriteString("No previous record found.\n")
	}
	if len(sb.String()) == 0 {
		sb.WriteString("No data available for the daily report.\n")
	}
	if len(failedHosts) > 0 {
		sb.WriteString("🛑 Failed to connect to the following accounts:\n")
		for _, host := range failedHosts {
			sb.WriteString(fmt.Sprintf("    - %s\n", host))
		}
	}
	logger.Debugf("Generated daily report:\n%s", sb.String())
	// Return the report as a string
	logger.Infof("Daily report generated successfully")
	// Ensure the report is not empty
	if sb.Len() == 0 {
		return "", fmt.Errorf("daily report is empty, no data available")
	}
	// Return the report string
	logger.Infof("Returning daily report with %d characters", sb.Len())
	return sb.String(), nil
}
