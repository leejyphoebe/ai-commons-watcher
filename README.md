# ai-commons

Script for running AI workloads on Aspire2A NSCC cluster.

## Pre-requisites

1. Git must be installed and configured in the login node
2. `go version` == 1.24.3

## Building
`slack/main.go`: `CGO_LDFLAGS="-lm" go build -o bin/slack_bot slack/main.go`