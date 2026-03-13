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
  diff        Show diff between local definition and deployed version
  versions    List application versions
  rollback    Rollback to a previous version

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

Most fields correspond to the [`version.CreateParams`](https://pkg.go.dev/github.com/sacloud/apprun-dedicated-api-go/apis/version#CreateParams) in the AppRun Dedicated API Go SDK.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cluster` | string | yes | Cluster name to deploy to (CLI only, not sent to API) |
| `name` | string | yes | Application name (CLI only, not sent to API) |
| `cpu` | int | | CPU cores |
| `memory` | int | | Memory in GB |
| `image` | string | | Container image (e.g., `"ghcr.io/example/app:v1.0.0"`) |
| `scalingMode` | string | | `"fixed"` or `"auto"` |
| `fixedScale` | int | | Number of instances (fixed scaling) |
| `minScale` | int | | Minimum instances (auto scaling) |
| `maxScale` | int | | Maximum instances (auto scaling) |
| `scaleInThreshold` | int | | Scale-in threshold percentage |
| `scaleOutThreshold` | int | | Scale-out threshold percentage |
| `cmd` | []string | | Override container command |
| `exposedPorts` | []ExposedPort | | Port exposure and LB routing settings |
| `env` | []EnvVar | | Environment variables (`key`, `value`, `secret`) |

### ExposedPort and Load Balancer Routing

In AppRun Dedicated, the Load Balancer (LB) is a shared resource within a cluster managed separately (e.g., via Terraform). The LB performs **host-based L7 routing** — it inspects the `Host` header of incoming HTTP requests and routes them to the appropriate application.

`exposedPorts` defines how your application connects to the LB:

```
Client → LB (host-based routing) → Application Container
         port 80/443                 targetPort 8080
         Host: app.example.com
```

- **`targetPort`** — The port your container listens on.
- **`loadBalancerPort`** — The LB port that should route traffic to this application. This must match an existing LB port in the cluster.
- **`host`** — The hostname(s) the LB uses to route requests to this application. **Required when `loadBalancerPort` is set.** Multiple applications can share the same LB port as long as they have different hostnames.

If your application does not need external access via the LB, you can omit `loadBalancerPort` and `host`, and only set `targetPort`.

Example: Expose an app on LB port 80 with hostname routing and health check:

```jsonnet
exposedPorts: [
  {
    targetPort: 8080,
    loadBalancerPort: 80,
    host: ["app.example.com"],
    healthCheck: {
      path: "/health",
      intervalSeconds: 10,
      timeoutSeconds: 5,
    },
  },
],
```

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

### diff

Show the difference between the local definition file and the currently deployed (active) version. Output is in colored unified diff format.

```console
$ apprun-dedicated-cli diff --app application.jsonnet
```

### versions

List all versions of the application as JSON. Each entry includes the version number, image, active node count, whether it is the active version, and the creation timestamp.

```console
$ apprun-dedicated-cli versions --app application.jsonnet
```

### rollback

Rollback to a previous version. By default, activates the latest existing version before the current active version. Use `--version` to specify a target version.

```console
# Rollback to the previous version
$ apprun-dedicated-cli rollback --app application.jsonnet

# Rollback to a specific version
$ apprun-dedicated-cli rollback --version 3 --app application.jsonnet
```

| Flag | Description |
|------|-------------|
| `--version` | Version number to rollback to (default: previous existing version) |

## License

MIT

## Author

fujiwara
