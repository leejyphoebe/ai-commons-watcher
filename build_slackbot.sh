#!/bin/bash

CGO_LDFLAGS="-lm" go build -o bin/slack_bot slack/main.go