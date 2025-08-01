#!/bin/bash

export $(xargs -a $HOME/.ai-commons/dockerhub.env)
docker pull meowth16/anyrepo:ai_commons_slack
