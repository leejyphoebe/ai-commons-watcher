FROM python:3.11-slim

# System deps for nbconvert PDF (xelatex + pandoc)
RUN apt-get update && apt-get install -y \
    build-essential \
    pandoc \
    texlive-xetex \
    texlive-fonts-recommended \
    texlive-plain-generic \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy watcher package and default Docker config
COPY watcher ./watcher
COPY config/config.docker.yaml ./config/config.docker.yaml

# Default entrypoint: use the Docker config
ENTRYPOINT ["python", "-m", "watcher.watcher", "--config", "/app/config/config.docker.yaml"]
