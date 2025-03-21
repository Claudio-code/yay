package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jguer/yay/v12/pkg/db/ialpm"
	"github.com/Jguer/yay/v12/pkg/dep"
	"github.com/Jguer/yay/v12/pkg/runtime"
	"github.com/Jguer/yay/v12/pkg/settings"
	"github.com/Jguer/yay/v12/pkg/settings/parser"
	"github.com/Jguer/yay/v12/pkg/text"

	"github.com/Jguer/aur/metadata"
	"github.com/leonelquinteros/gotext"
	"github.com/pkg/errors"
)

func handleCmd(logger *text.Logger) error {
	cfg, err := settings.NewConfig(logger, settings.GetConfigPath(), "")
	if err != nil {
		return err
	}

	cmdArgs := parser.MakeArguments()
	if errP := cfg.ParseCommandLine(cmdArgs, logger); errP != nil {
		return errP
	}

	run, err := runtime.NewRuntime(cfg, cmdArgs, "1.0.0")
	if err != nil {
		return err
	}

	dbExecutor, err := ialpm.NewExecutor(run.PacmanConf, logger)
	if err != nil {
		return err
	}

	aurCache, err := metadata.New(
		metadata.WithCacheFilePath(
			filepath.Join(cfg.BuildDir, "aur.json")))
	if err != nil {
		return errors.Wrap(err, gotext.Get("failed to retrieve aur Cache"))
	}

	grapher := dep.NewGrapher(dbExecutor, aurCache, true, settings.NoConfirm,
		cmdArgs.ExistsDouble("d", "nodeps"), false, false,
		run.Logger.Child("grapher"))

	return graphPackage(context.Background(), grapher, cmdArgs.Targets)
}

func main() {
	fallbackLog := text.NewLogger(os.Stdout, os.Stderr, os.Stdin, false, "fallback")
	if err := handleCmd(fallbackLog); err != nil {
		fallbackLog.Errorln(err)
		os.Exit(1)
	}
}

func graphPackage(
	ctx context.Context,
	grapher *dep.Grapher,
	targets []string,
) error {
	if len(targets) != 1 {
		return errors.New(gotext.Get("only one target is allowed"))
	}

	graph, err := grapher.GraphFromAUR(ctx, nil, []string{targets[0]})
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, graph.String())
	fmt.Fprintln(os.Stdout, "\nlayers map\n", graph.TopoSortedLayerMap(nil))

	return nil
}
