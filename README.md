# apprun-dedicated-cli

A CLI tool for deploying applications to [Sakura Cloud AppRun Dedicated](https://manual.sakura.ad.jp/cloud/apprun/dedicated/index.html) (専有型).

Inspired by [ecspresso](https://github.com/kayac/ecspresso) for Amazon ECS, this tool focuses exclusively on application deployment (create/update application versions), while infrastructure (clusters, ASGs, load balancers) is expected to be managed separately (e.g., via Terraform).

## Install

### Binary releases

Download from [GitHub Releases](https://github.com/fujiwara/apprun-dedicated-cli/releases).

### Homebrew

```console
$ brew install fujiwara/tap/apprun-dedicated-cli
```

### Go install

```console
$ go install github.com/fujiwara/apprun-dedicated-cli/cmd/apprun-dedicated-cli@latest
```

### GitHub Actions

```yaml
- uses: fujiwara/apprun-dedicated-cli@v0
  with:
    version: v0.1.0
```

You can also specify a version file or run commands directly:

```yaml
- uses: fujiwara/apprun-dedicated-cli@v0
  with:
    version: v0.1.0
    args: "deploy --app application.jsonnet"
```

| Input | Description |
|-------|-------------|
| `version` | Version to install (e.g., `v0.1.0`) |
| `version-file` | File containing the version (alternative to `version`) |
| `args` | Arguments to run after installation (optional) |

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
  delete         Delete application
  deactivate     Deactivate application (stop without deleting)
  render         Render definition file
  status         Show status of application
  diff           Show diff between local definition and deployed version
  versions       List application versions
  rollback       Rollback to a previous version
  containers     Show container status of application
  cluster        Show cluster information
  load-balancer  Show load balancer information (alias: lb)
  certificate    Show certificate information
  asg            Show auto scaling group information
  worker-node    Show worker node information

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
| `registryUsername` | string | | Container registry username |
| `registryPassword` | string | | Container registry password |
| `exposedPorts` | []ExposedPort | | Port exposure and LB routing settings |
| `env` | []EnvVar | | Environment variables (`key`, `value`, `secret`) |

### Private Container Registry

To pull images from a private registry, set `registryUsername` and `registryPassword` in the definition file. Use jsonnet native functions to avoid hardcoding credentials:

```jsonnet
{
  image: "registry.example.com/my-app:v1.0.0",
  registryUsername: std.native("env")("REGISTRY_USERNAME", ""),
  registryPassword: std.native("env")("REGISTRY_PASSWORD", ""),
  // ...
}
```

When `registryPassword` is set to a non-empty string, the password is updated on each deploy. When omitted or empty (`""`), the existing password is kept for subsequent deploys.

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

- **`useLetsEncrypt`** — Enable automatic TLS certificate provisioning via Let's Encrypt (see [HTTPS with Let's Encrypt](#https-with-lets-encrypt) below).

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

### HTTPS with Let's Encrypt

To serve your application over HTTPS using Let's Encrypt automatic certificates:

1. **Cluster prerequisites** — The cluster must be created with the following settings:
   - **Let's Encrypt enabled** with a valid email address
   - **LB port 80 (HTTP)** — Required for Let's Encrypt HTTP-01 challenge
   - **LB port 443 (HTTPS)** — For serving HTTPS traffic

   Note: LB ports can only be configured at cluster creation time and cannot be added later.

2. **DNS setup** — Point your domain to the LB node's IP address. You can find the IP using the `load-balancer` command:
   ```console
   $ apprun-dedicated-cli load-balancer --app application.jsonnet
   ```

3. **Application definition** — Set `loadBalancerPort` to 443, specify your hostname in `host`, and enable `useLetsEncrypt`:
   ```jsonnet
   exposedPorts: [
     {
       targetPort: 8080,
       loadBalancerPort: 443,
       host: ["app.example.com"],
       useLetsEncrypt: true,
     },
   ],
   ```

4. **Deploy** — Run `apprun-dedicated-cli deploy` to apply the configuration. Certificate provisioning by Let's Encrypt may take a few minutes.

## Stopping and Resuming Applications

AppRun Dedicated does not support scaling to zero instances. To stop an application, use `deactivate` to remove the active version, which stops all running containers.

### Stop an application

```console
$ apprun-dedicated-cli deactivate --app application.jsonnet
```

The command deactivates the application and waits for it to fully stop. The application and all its versions are preserved.

### Resume a stopped application

Use `rollback` to reactivate the latest (or a specific) version:

```console
# Reactivate the latest version
$ apprun-dedicated-cli rollback --app application.jsonnet

# Reactivate a specific version
$ apprun-dedicated-cli rollback --target 3 --app application.jsonnet
```

Alternatively, `deploy` will create a new version from the current definition file and activate it.

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

Deploy the application. Creates the application if it does not exist, then creates a new version and activates it (idempotent). By default, waits for all containers to be running before returning.

```console
$ apprun-dedicated-cli deploy --app application.jsonnet

# Skip waiting for deployment to complete
$ apprun-dedicated-cli deploy --no-wait --app application.jsonnet
```

| Flag | Description |
|------|-------------|
| `--no-wait` | Skip waiting for deployment to complete (default: `--wait`) |

### delete

Delete an application. Prompts for confirmation unless `--force` is specified.

```console
$ apprun-dedicated-cli delete --app application.jsonnet

# Skip confirmation
$ apprun-dedicated-cli delete --force --app application.jsonnet
```

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |

### deactivate

Deactivate an application. This stops the application without deleting it — the application and its versions are preserved. By default, waits for the application to fully stop before returning. Prompts for confirmation unless `--force` is specified.

```console
$ apprun-dedicated-cli deactivate --app application.jsonnet

# Skip confirmation
$ apprun-dedicated-cli deactivate --force --app application.jsonnet

# Return immediately without waiting for stop
$ apprun-dedicated-cli deactivate --no-wait --app application.jsonnet
```

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |
| `--no-wait` | Skip waiting for application to stop (default: `--wait`) |

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

Rollback to a previous version, or reactivate a deactivated application.

- If the application is active, activates the latest version before the current one (or the version specified by `--target`).
- If the application is deactivated, activates the latest version (or the version specified by `--target`).

```console
# Rollback to the previous version
$ apprun-dedicated-cli rollback --app application.jsonnet

# Rollback to a specific version
$ apprun-dedicated-cli rollback --target 3 --app application.jsonnet

# Reactivate a deactivated application
$ apprun-dedicated-cli rollback --app application.jsonnet
```

| Flag | Description |
|------|-------------|
| `--target` | Version number to activate (default: previous version, or latest if deactivated) |
| `--no-wait` | Skip waiting for deployment to complete (default: `--wait`) |

### containers

Show the current container placement and status for the application as JSON. Each entry includes node ID, container states, and desired container configuration.

```console
$ apprun-dedicated-cli containers --app application.jsonnet
```

### cluster

Show cluster information as JSON. The cluster name is read from the application definition file.

```console
$ apprun-dedicated-cli cluster --app application.jsonnet
```

### load-balancer

Show load balancer information as JSON. Lists all load balancers in the cluster, including node details with IP addresses. Also available as `lb`.

```console
$ apprun-dedicated-cli load-balancer --app application.jsonnet
$ apprun-dedicated-cli lb --app application.jsonnet
```

### certificate

Show certificate information as JSON. Lists all certificates in the cluster.

```console
$ apprun-dedicated-cli certificate --app application.jsonnet
```

### asg

Show auto scaling group information as JSON. Lists all auto scaling groups in the cluster.

```console
$ apprun-dedicated-cli asg --app application.jsonnet
```

### worker-node

Show worker node information as JSON. Lists all worker nodes across auto scaling groups in the cluster.

```console
$ apprun-dedicated-cli worker-node --app application.jsonnet
```

## License

MIT

## Author

fujiwara
