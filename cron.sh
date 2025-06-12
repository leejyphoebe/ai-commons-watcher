#!/usr/bin/bash

# Function to add a cron job
add_cron_job() {
  local schedule="$1"
  local command="$2"
  local log_file="$3"

  if [ -z "$schedule" ] || [ -z "$command" ]; then
    echo "Error: Schedule and command are required."
    return 1
  fi

  if [ -z "$log_file" ]; then
    log_file="/dev/null" # Default to /dev/null if no log file is specified
  fi

  # Check if the cron job already exists
  if crontab -l | grep -q "$schedule.*$command"; then
      echo "Cron job already exists: $schedule $command"
      return 0
  fi

  # Add the cron job
  (crontab -l ; echo "$schedule $command >> $log_file 2>&1") | crontab -
  echo "Cron job added: $schedule $command >> $log_file 2>&1"
  return 0
}

project_dir="${HOME}/ai-commons"
cmd="go run $project_dir/slack/main.go"

# Add a cron job to run a script at 9am on weekdays and log output
add_cron_job "0 9 * * Mon-Fri" "$cmd --title 'NSCC Usage AM Report'" "$project_dir/logs/cron_nscc_am.log"

# Add a cron job to run a command at 6pm on weekdays and log output
add_cron_job "0 18 * * Mon-Fri" "$cmd --title 'NSCC Usage PM Report'" "$project_dir/logs/cron_nscc_pm.log"

# List existing cron jobs
echo "Current cron jobs:"
crontab -l
