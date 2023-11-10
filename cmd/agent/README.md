# cmd/agent

## Description

This is a agent which send metrics to server. 

## Run

```bash
$ go run ./cmd/agent
```

### Possible flags

- `-a` - Host of the server which will collect metrics. Default: `localhost:8080`. Alias for `ADDRESS` in env.
- `-p` - Polling interval in seconds. Default: `5`. Alias for `POLL_INTERVAL` in env.
- `-r` - Reporting interval in seconds. Default: `2`. Alias for `REPORT_INTERVAL` in env.

## Test

```bash
$ go test ./internal/agent/...
```git 