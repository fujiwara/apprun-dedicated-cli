# Manual Verification Steps

## Prerequisites

- Go 1.21+
- Environment variables set:
  - `SAKURA_ACCESS_TOKEN`
  - `SAKURA_ACCESS_TOKEN_SECRET`
- An existing AppRun Dedicated cluster named `default` with an application named `printenv` in zone `is1c`

## Build

```bash
go build -o apprun-dedicated-cli ./cmd/apprun-dedicated-cli/
```

## 1. Init: Generate config and application definition from existing resources

```bash
./apprun-dedicated-cli init --cluster default --application printenv -o /tmp/apprun-test
```

**Expected**: Two files are generated in `/tmp/apprun-test/`:
- `config.jsonnet` — config file with cluster name, application name, and pointer to definition file
- `application.jsonnet` — application definition (CPU, memory, image, ports, env vars, etc.)

Inspect the output:

```bash
cat /tmp/apprun-test/config.jsonnet
cat /tmp/apprun-test/application.jsonnet
```

## 2. Deploy: Create a new version and activate it

Edit the application definition (e.g., change an environment variable value) to confirm that deploy creates a new version:

```bash
vi /tmp/apprun-test/application.jsonnet
```

Run deploy:

```bash
./apprun-dedicated-cli -c /tmp/apprun-test/config.jsonnet deploy
```

**Expected**: A new version is created and activated. Log output shows the new version number.

## 3. Status: Verify the deployed version

```bash
./apprun-dedicated-cli -c /tmp/apprun-test/config.jsonnet status
```

**Expected**: JSON output showing cluster name, application info (including `active_version`), and `active_version_detail` with the image, CPU, memory, etc.

## 4. Render: Preview definition file output

```bash
./apprun-dedicated-cli -c /tmp/apprun-test/config.jsonnet render
```

**Expected**: Rendered JSON output of the application definition to stdout.

## 5. Debug logging

Add `--debug` to any command to see detailed logs:

```bash
./apprun-dedicated-cli -c /tmp/apprun-test/config.jsonnet --debug status
```

Add `--log-format json` for JSON-formatted logs:

```bash
./apprun-dedicated-cli -c /tmp/apprun-test/config.jsonnet --log-format json status
```

## Cleanup

```bash
rm -rf /tmp/apprun-test
```
