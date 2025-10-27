package main

import (
	"context"
	"fmt"

	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Show current session status",
		Action: StatusAction,
	}
}

func StatusAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

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
