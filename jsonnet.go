package cli

import (
	"context"
	"log/slog"

	sscli "github.com/fujiwara/sakura-secrets-cli"
	"github.com/fujiwara/tfstate-lookup/tfstate"
	"github.com/google/go-jsonnet"
)

func (c *CLI) setupVM(ctx context.Context) error {
	nativeFuncs := defaultJsonnetNativeFuncs()

	// load secretmanager functions
	secretFunc := sscli.SecretNativeFunction(ctx)
	nativeFuncs = append(nativeFuncs, secretFunc)

	// load tfstate functions
	if c.TFState != "" {
		slog.Debug("loading tfstate", "url", c.TFState)
		lookup, err := tfstate.ReadURL(ctx, c.TFState)
		if err != nil {
			return err
		}
		nativeFuncs = append(nativeFuncs, lookup.JsonnetNativeFuncs(ctx)...)
	}

	c.loader.AddFunctions(nativeFuncs...)
	return nil
}

func defaultJsonnetNativeFuncs() []*jsonnet.NativeFunction {
	return []*jsonnet.NativeFunction{}
}
