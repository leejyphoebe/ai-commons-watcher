.PHONY: build_slack_bot run_slack_bot build run

CMD_DIR := $(PWD)/cmd
OUT_DIR := $(PWD)/bin

print_path_instruction:
	@echo "To use built tools from anywhere, add to PATH:"
	@echo "    export PATH=\$$PATH:$(OUT_DIR)"
	@echo

build_slack_bot: print_path_instruction
	CGO_LDFLAGS="-lm" go build -o $(OUT_DIR)/slack_bot $(CMD_DIR)/slack/main.go 
	@echo
	@echo "✅ Slack bot built successfully at: $(OUT_DIR)/slack_bot"
	@echo
	@echo "Run the bot using:"
	@echo "    slack_bot --config /path/to/config.yaml --title 'Message Title'"
	@echo
	@echo "Make sure to set the following environment variables as needed before running the bot."
	@echo "SLACK_AUTH_TOKEN, SLACK_CHANNEL_ID, BITWARDEN_ACCESS_TOKEN, BITWARDEN_CLIENT_ID, BITWARDEN_CLIENT_SECRET, BITWARDEN_ORG_ID"

run_slack_bot: print_path_instruction
	@CONFIG=$${config:-}; \
	if [ -z "$$CONFIG" ]; then CONFIG=example/config.yaml; fi; \
	echo "🧪 Running Slack bot with config: $$CONFIG"; \
	TITLE=$${title:-}; \
	if [ -z "$$TITLE" ]; then TITLE="NSCC Usage Report"; fi; \
	SECRET=$${secret:-}; \
	if [ -z "$$SECRET" ]; then SECRET=example/secret/slack.env; fi; \
	export $(cat $$SECRET | xargs) \
	slack_bot --config $$CONFIG --title $$TITLE && \
	echo "✅ Slack bot ran successfully" || \
	(echo "❌ Failed to run Slack bot with config: $$CONFIG" && exit 1)

cron_slack_bot:
	./slack_scripts/cron.sh

build: print_path_instruction
	go build -o $(OUT_DIR)/ai-commons $(CMD_DIR)/main
	@echo
	@echo "✅ AI Commons built successfully at: $(OUT_DIR)/ai-commons"
	@echo
	@echo "You can now run the cli application using:"
	@echo "    ai-commons --config /path/to/config.yaml --title 'Your Title'"
	@echo 
	@echo "Make sure to set the following environment variables as needed before running the AI Commons."
	@echo "BITWARDEN_ACCESS_TOKEN, BITWARDEN_CLIENT_ID, BITWARDEN_CLIENT_SECRET, BITWARDEN_ORG_ID"

run: print_path_instruction
	@CONFIG=$${config:-}; \
	if [ -z "$$CONFIG" ]; then CONFIG=example/config.yaml; fi; \
	echo "🧪 Running experiment with config: $$CONFIG"; \
	ai-commons run --config $$CONFIG	&& \
	echo "✅ Experiment ran successfully" || \
	(echo "❌ Failed to run experiment with config: $$CONFIG" && exit 1)
