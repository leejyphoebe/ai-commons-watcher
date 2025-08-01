#!/bin/bash

title=$1
docker run --rm \
  -v "$HOME/.ai-commons:/app/.ai-commons" \
  -v "$HOME/.ssh:/app/.ssh" \
  --env-file="$HOME/.ai-commons/slack.env" \
  ${DOCKER_IMAGE:-meowth16/anyrepo:ai_commons_slack} \
    ./slack_bot \
    --config /app/.ai-commons/config.yaml \
    --title "$title"
