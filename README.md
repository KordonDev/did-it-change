# Did It Change

A service written in Go that monitors URLs for content changes and reports their status.

## Features

- Monitor multiple endpoints for content changes
- Configure check intervals and fail thresholds via YAML
- REST API to check the status of monitors
- Automatic detection of unchanged content

## Installation

You need Go 1.18 or higher installed.

```bash
# Install dependencies
go mod tidy
```

## Configuration

Create or modify the configuration file at `config/monitors.yaml`:

```yaml
monitors:
  - name: example-monitor
    endpoint: https://example.com/api
    checkInterval: 300  # seconds between checks
    failThreshold: 3    # consecutive unchanged checks to mark as fail
```

## Usage

Start the service:

```bash
go run *.go
```

Or build and run:

```bash
go build
./did-it-change
```

### Docker

You can also run the application using Docker:

```bash
# Build the Docker image
docker build -t did-it-change .

# Run the container
docker run -p 8080:8080 -v $(pwd)/config:/app/config did-it-change
```

Or use the pre-built image from GitHub Container Registry:

```bash
docker pull ghcr.io/kordondev/did-it-change:latest
docker run -p 8080:8080 -v $(pwd)/config:/app/config ghcr.io/kordondev/did-it-change:latest
```

## API Endpoints

- `GET /api/monitors` - Get all monitors and their statuses
- `GET /api/monitors/:name` - Get status for a specific monitor
- `GET /health` - Health check endpoint

## How it works

1. The service loads all monitors from the configuration file
2. Each endpoint is checked at the specified interval
3. Content from the endpoint is hashed and compared to the previous hash
4. If the content remains unchanged for `failThreshold` consecutive checks, the status is set to `fail`
5. If the content changes after being marked as `fail`, the status is set to `success`
