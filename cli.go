package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	armed "github.com/fujiwara/jsonnet-armed"
	"github.com/fujiwara/sloghandler"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
)

type CLI struct {
	// Commands
	Init     InitOption     `cmd:"" help:"Initialize definition files from existing resources"`
	Deploy   DeployOption   `cmd:"" help:"Deploy application"`
	Diff     DiffOption     `cmd:"" help:"Show diff of definitions"`
	Render   RenderOption   `cmd:"" help:"Render definition files"`
	Status   StatusOption   `cmd:"" help:"Show status of application"`
	Versions VersionsOption `cmd:"" help:"List application versions"`
	Rollback RollbackOption `cmd:"" help:"Rollback to previous version"`

	// Global flags
	Config    string           `name:"config" short:"c" help:"Path to config file" default:"config.jsonnet" env:"APPRUN_DEDICATED_CONFIG"`
	Debug     bool             `help:"Enable debug mode" env:"DEBUG"`
	LogFormat string           `name:"log-format" help:"Log format (text or json)" default:"text" enum:"text,json" env:"APPRUN_DEDICATED_LOG_FORMAT"`
	TFState   string           `name:"tfstate" help:"URL to terraform.tfstate" env:"APPRUN_DEDICATED_TFSTATE"`
	Version   kong.VersionFlag `short:"v" help:"Show version and exit."`

	// internal
	sc     *saclient.Client
	client *v1.Client
	loader *armed.CLI
	config *Config
}

func (c *CLI) Run(ctx context.Context) error {
	k := kong.Parse(c, kong.Vars{"version": fmt.Sprintf("apprun-dedicated-cli %s", Version)})

	c.setupLogger()

	if err := c.setupVM(ctx); err != nil {
		return err
	}

	if err := c.setupClient(); err != nil {
		return err
	}

	// init generates config file, so skip loadConfig
	if k.Command() == "init" {
		return c.runInit(ctx)
	}

	if err := c.loadConfig(ctx); err != nil {
		return err
	}

	var err error
	switch k.Command() {
	case "deploy":
		err = c.runDeploy(ctx)
	case "diff":
		err = c.runDiff(ctx)
	case "render":
		err = c.runRender(ctx)
	case "status":
		err = c.runStatus(ctx)
	case "versions":
		err = c.runVersions(ctx)
	case "rollback":
		err = c.runRollback(ctx)
	default:
		err = fmt.Errorf("unknown command: %s", k.Command())
	}
	return err
}

func (c *CLI) setupLogger() {
	level := slog.LevelInfo
	if c.Debug {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	switch c.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	default:
		handler = sloghandler.NewLogHandler(os.Stderr, &sloghandler.HandlerOptions{
			HandlerOptions: slog.HandlerOptions{Level: level},
			Color:          true,
		})
	}
	slog.SetDefault(slog.New(handler))
}

func (c *CLI) setupClient() error {
	if err := c.sc.SetEnviron(os.Environ()); err != nil {
		return fmt.Errorf("failed to set environ: %w", err)
	}
	if err := c.sc.Populate(); err != nil {
		return fmt.Errorf("failed to populate client: %w", err)
	}

	var err error
	c.client, err = newClient(c.sc)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	return nil
}

// Command option types
type InitOption struct {
	Cluster     string `help:"Cluster name" required:""`
	Application string `help:"Application name" required:""`
	OutputDir   string `name:"output-dir" short:"o" help:"Output directory for generated files" default:"."`
}
type DeployOption struct{}
type DiffOption struct{}
type RenderOption struct{}
type StatusOption struct{}
type VersionsOption struct{}
type RollbackOption struct{}

// Placeholder command implementations (TODO)
func (c *CLI) runDiff(ctx context.Context) error     { return fmt.Errorf("not implemented yet") }
func (c *CLI) runVersions(ctx context.Context) error { return fmt.Errorf("not implemented yet") }
func (c *CLI) runRollback(ctx context.Context) error { return fmt.Errorf("not implemented yet") }
