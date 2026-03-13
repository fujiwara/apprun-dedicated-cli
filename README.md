# apprun-dedicated-cli

A CLI tool for deploying applications to [Sakura Cloud AppRun Dedicated](https://manual.sakura.ad.jp/cloud/apprun/dedicated/index.html) (専有型).

Inspired by [ecspresso](https://github.com/kayac/ecspresso) for Amazon ECS, this tool focuses exclusively on application deployment (create/update application versions), while infrastructure (clusters, ASGs, load balancers) is expected to be managed separately (e.g., via Terraform).

## Install

```console
$ go install github.com/fujiwara/apprun-dedicated-cli/cmd/apprun-dedicated-cli@latest
```

## Authentication

Set the following environment variables:

```
SAKURA_ACCESS_TOKEN=<your access token>
SAKURA_ACCESS_TOKEN_SECRET=<your access token secret>
```

## Usage

```
Usage: apprun-dedicated-cli <command> [flags]

Commands:
  init        Initialize definition file from existing resources
  deploy      Deploy application
  render      Render definition file
  status      Show status of application
  diff        Show diff of definitions (not yet implemented)
  versions    List application versions (not yet implemented)
  rollback    Rollback to previous version (not yet implemented)

Global Flags:
  --app=STRING          Path to application definition file (default: "application.jsonnet", env: APPRUN_DEDICATED_APP)
  --debug               Enable debug mode (env: DEBUG)
  --log-format=STRING   Log format: text or json (default: "text", env: APPRUN_DEDICATED_LOG_FORMAT)
  --tfstate=STRING      URL to terraform.tfstate (env: APPRUN_DEDICATED_TFSTATE)
  -v, --version         Show version and exit
```

## Application Definition File

The application definition is written in [Jsonnet](https://jsonnet.org/) format. The JSON field names match the AppRun Dedicated API (camelCase) for forward compatibility.

```jsonnet
{
  cluster: "my-cluster",
  name: "my-app",
  cpu: 1,
  memory: 1,
  image: "ghcr.io/example/app:v1.0.0",
  exposedPorts: [
    {
      targetPort: 8080,
      loadBalancerPort: 80,
      host: ["app.example.com"],
    },
  ],
  env: [
    { key: "FOO", value: "bar" },
  ],
}
```

### Definition Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cluster` | string | yes | Cluster name |
| `name` | string | yes | Application name |
| `cpu` | int | | CPU cores |
| `memory` | int | | Memory in GB |
| `image` | string | | Container image (e.g., `"ghcr.io/example/app:v1.0.0"`) |
| `scalingMode` | string | | `"fixed"` or `"auto"` |
| `fixedScale` | int | | Number of instances (when scalingMode is `"fixed"`) |
| `minScale` | int | | Minimum instances (when scalingMode is `"auto"`) |
| `maxScale` | int | | Maximum instances (when scalingMode is `"auto"`) |
| `scaleInThreshold` | int | | Scale-in threshold percentage |
| `scaleOutThreshold` | int | | Scale-out threshold percentage |
| `cmd` | []string | | Override container command |
| `exposedPorts` | []ExposedPort | | Port exposure and LB routing settings |
| `env` | []EnvVar | | Environment variables |

### ExposedPort

| Field | Type | Description |
|-------|------|-------------|
| `targetPort` | int | Container port |
| `loadBalancerPort` | int | Load balancer port (requires `host`) |
| `useLetsEncrypt` | bool | Enable Let's Encrypt TLS |
| `host` | []string | Hostnames for L7 routing |
| `healthCheck` | HealthCheck | Health check settings |

### HealthCheck

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Health check path |
| `intervalSeconds` | int | Check interval |
| `timeoutSeconds` | int | Check timeout |

### EnvVar

| Field | Type | Description |
|-------|------|-------------|
| `key` | string | Variable name |
| `value` | string | Variable value |
| `secret` | bool | Whether the value is a secret |

## Jsonnet Functions

The definition file is evaluated with [jsonnet-armed](https://github.com/fujiwara/jsonnet-armed), which provides the following native functions:

- `std.native("env")("KEY", "default")` — Read environment variable
- `std.native("must_env")("KEY")` — Read environment variable (error if not set)
- `std.native("secret")("KEY")` — Read from [Sakura Cloud Secret Manager](https://github.com/fujiwara/sakura-secrets-cli)

When `--tfstate` is specified, [tfstate-lookup](https://github.com/fujiwara/tfstate-lookup) functions are also available:

- `std.native("tfstate")("resource.type.name.attr")` — Look up Terraform state values

## Commands

### init

Generate an application definition file from an existing application.

```console
$ apprun-dedicated-cli init --cluster my-cluster --application my-app -o ./myapp/
```

| Flag | Description |
|------|-------------|
| `--cluster` | Cluster name (required) |
| `--application` | Application name (required) |
| `-o, --output-dir` | Output directory (default: `.`) |

### deploy

Deploy the application. Creates the application if it does not exist, then creates a new version and activates it (idempotent).

```console
$ apprun-dedicated-cli deploy --app application.jsonnet
```

### status

Show the current status of the application as JSON.

```console
$ apprun-dedicated-cli status --app application.jsonnet
```

### render

Render and output the evaluated application definition as JSON.

```console
$ apprun-dedicated-cli render --app application.jsonnet
```

## License

MIT

## Author

fujiwara
