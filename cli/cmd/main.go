package main

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func main() {
	ui.InitLogger(log.InfoLevel)

	logger := ui.GetLogger()
	ctx := context.Background()
	reg := registry.Get()

	if err := reg.Init(ctx); err != nil {
		logger.Fatalf("Failed to initialize registry %v", err)
	}
	defer reg.Close()

	app := &cli.Command{
		Name:    "skycli",
		Usage:   "A companion CLI tool for your Bluesky feed ecosystem",
		Version: "0.1.0",
		Commands: []*cli.Command{
			SetupCommand(), LoginCommand(), StatusCommand(),
			ListCommand(), ViewCommand(), ExportCommand(),
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		logger.Fatalf("Command failed with error: %v", err)
	}
}
