#!/usr/bin/bash

title=$1
export $(cat /home/meow/vhive/nscc/ai-commons/.env | xargs)
/home/meow/vhive/nscc/ai-commons/bin/slack_bot --config /home/meow/vhive/nscc/ai-commons/config.yaml --title "$1"
