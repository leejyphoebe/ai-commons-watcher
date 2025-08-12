#!/bin/bash

set -a
source $HOME/.ai-commons/dockerhub.env
set +a
docker pull meowth16/anyrepo:ai_commons_slack
