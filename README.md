# ai-commons
Tools for running AI workloads on Aspire2A NSCC cluster and CCDS GPU clusters.

## Building Docker Images
1. Get the latest commit hash from your specific branch on GitHub
```bash
COMMIT_HASH=$(git ls-remote ${VLLM_REPO_URL} ${VLLM_BRANCH} | cut -f1)
```

2. Pass it to the build command
```bash
docker build \
  --build-arg PYTHON_VERSION=3.12 \
  --build-arg VLLM_REPO_URL=https://github.com/googlercolin/vllm.git \
  --build-arg VLLM_BRANCH=colin_v0.9.0.1 \
  --build-arg CACHE_INVALIDATOR=${COMMIT_HASH} \
  -f Dockerfile.nscc \
  -t your-image-name:latest \
  .
```

## Building the Slack Bot

### Pre-requisites

1. Git must be installed and configured in the login node
2. `go version` == 1.24.3

### Building
`slack/main.go`: `CGO_LDFLAGS="-lm" go build -o bin/slack_bot slack/main.go`
