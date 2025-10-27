package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func main() {
	ui.InitLogger(log.InfoLevel)
	logger := ui.GetLogger()

	ctx := context.Background()
	reg := registry.Get()

	if err := reg.Init(ctx); err != nil {
		logger.Fatal("Failed to initialize registry", "error", err)
	}
	defer reg.Close()

	app := &cli.Command{
		Name:    "skycli",
		Usage:   "A companion CLI tool for your Bluesky feed ecosystem",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "Authenticate with Bluesky",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "handle",
						Aliases:  []string{"u"},
						Usage:    "Your Bluesky handle (e.g., @user.bsky.social)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Your app password",
						Required: true,
					},
				},
				Action: loginAction,
			},
			{
				Name:   "status",
				Usage:  "Show current session status",
				Action: statusAction,
			},
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		logger.Fatal("Command failed", "error", err)
	}
}

func loginAction(ctx context.Context, cmd *cli.Command) error {
	logger := ui.GetLogger()
	reg := registry.Get()

	handle := cmd.String("handle")
	password := cmd.String("password")

	logger.Info("Authenticating with Bluesky", "handle", handle)

	service, err := reg.GetService()
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	credentials := map[string]string{
		"identifier": handle,
		"password":   password,
	}

	if err := service.Authenticate(ctx, credentials); err != nil {
		logger.Error("Authentication failed", "error", err)
		return err
	}

	sessionRepo, err := reg.GetSessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository: %w", err)
	}

	if err := sessionRepo.UpdateTokens(ctx, service.GetAccessToken(), service.GetRefreshToken()); err != nil {
		logger.Warn("Failed to save session tokens", "error", err)
	}

	ui.Successln("Successfully authenticated as %s", handle)
	return nil
}

func statusAction(ctx context.Context, cmd *cli.Command) error {
	logger := ui.GetLogger()
	reg := registry.Get()

	sessionRepo, err := reg.GetSessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository: %w", err)
	}

	if !sessionRepo.HasValidSession(ctx) {
		ui.Infoln("Not authenticated. Run 'skycli login' to authenticate.")
		return nil
	}

	session, err := sessionRepo.List(ctx)
	if err != nil {
		logger.Error("Failed to get session", "error", err)
		return err
	}

	if len(session) > 0 {
		if s, ok := session[0].(*store.SessionModel); ok {
			ui.Titleln("Session Status")
			ui.Infoln("Handle: %s", s.Handle)
			ui.Infoln("Service: %s", s.ServiceURL)
			ui.Successln("Authenticated")
		}
	}

	return nil
}
