# CLAUDE.md

## Project Overview

CLI tool for deploying applications to [Sakura Cloud AppRun Dedicated](https://manual.sakura.ad.jp/cloud/apprun/dedicated/index.html) (専有型). Inspired by [ecspresso](https://github.com/kayac/ecspresso).

## Build & Test

```sh
# Build
make apprun-dedicated-cli

# Test
make test          # or: go test -v ./...

# Install
make install

# Release build (snapshot)
make dist
```

## Key Conventions

- Go module: `github.com/fujiwara/apprun-dedicated-cli`
- Entry point: `cmd/apprun-dedicated-cli/main.go`
- Application config format: Jsonnet (`application.jsonnet`)
- JSON field names use camelCase (matching AppRun Dedicated API)
- CLI framework: `alecthomas/kong`
- API SDK: `github.com/sacloud/apprun-dedicated-api-go`
- Authentication: `SAKURA_ACCESS_TOKEN` and `SAKURA_ACCESS_TOKEN_SECRET` env vars

## Code Structure

- `cli.go` — CLI structure and command routing
- `client.go` — API client initialization
- `definition.go` — ApplicationDefinition struct & validation
- `convert.go` — Conversion between definition & SDK types
- `jsonnet.go` — Jsonnet VM setup with native functions
- `deploy.go`, `delete.go`, `deactivate.go`, `diff.go`, `init.go`, `status.go`, `render.go`, `versions.go`, `rollback.go` — Application command implementations
- `cluster.go`, `loadbalancer.go`, `certificate.go`, `autoscalinggroup.go`, `workernode.go` — Cluster resource info commands

## Development Notes

- Run `go fmt ./...` before committing
- Binary name: `apprun-dedicated-cli`
- When adding or changing features, always update README.md accordingly
