# cmd/server

## Description

This is a server for collecting metrics. 

Now used only memory storage.

## Run

```bash
$ go run ./cmd/server
```

### Possible flags

- `-a` - Host of the server. Default: `localhost:8080`. Alias for `ADDRESS` in env.
- `-d` - Postgres DSN. Default: `''`. Alias for `DATABASE_DSN` in env.
- `-f` - File storage path. Default: `/tmp/metrics-db.json`. Alias for `FILE_STORAGE_PATH` in env.
- `-i` - Store interval in seconds. Default: `300`. Alias for `STORE_INTERVAL` in env.
- `-r` - Restore from file. Default: `true`. Alias for `RESTORE` in env.
- `-k` - Key for hash ecnoding. Default empty. Alias for `KEY` in env.

## Test

```bash
$ go test ./internal/server/...
```