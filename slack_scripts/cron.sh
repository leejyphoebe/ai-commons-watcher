#!/usr/bin/bash

project_dir=$PWD
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

add_run_slackbot() {
  local script_path="$project_dir/bin/run_slackbot.sh"
  echo "#!/usr/bin/bash" > "$script_path"
  echo "" >> "$script_path"
  echo "title=\$1" >> "$script_path"
  echo "export \$(cat $project_dir/example/secret/slack.env | xargs)" >> "$script_path"
  echo "$project_dir/bin/slack_bot --config $project_dir/example/config.yaml --title \"\$1\"" >> "$script_path"
  chmod +x "$script_path"
  echo "Created run_slackbot script at $script_path"
}

# execute run_slackbot.sh
cmd="$project_dir/bin/run_slackbot.sh"
date_cmd=$(which date)
# Ensure the logs directory exists
mkdir -p "$HOME/logs"
add_run_slackbot

# Add a cron job to run a script at 9am on weekdays and log output
add_cron_job "0 1 * * Mon-Fri" "$cmd 'NSCC Usage AM Report'" "$project_dir/logs/cron_nscc_am_\`$date_cmd +\%Y\%m\%d\`.log"

# Add a cron job to run a command at 6pm on weekdays and log output
add_cron_job "0 9 * * Mon-Fri" "$cmd 'NSCC Usage PM Report'" "$project_dir/logs/cron_nscc_pm_\`$date_cmd +\%Y\%m\%d\`.log"

# List existing cron jobs
echo "Current cron jobs:"
crontab -l
