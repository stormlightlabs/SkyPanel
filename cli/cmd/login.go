package main

import (
	"context"
	"fmt"

	"github.com/stormlightlabs/skypanel/cli/internal/imports"
	"github.com/stormlightlabs/skypanel/cli/internal/registry"
	"github.com/stormlightlabs/skypanel/cli/internal/setup"
	"github.com/stormlightlabs/skypanel/cli/internal/store"
	"github.com/stormlightlabs/skypanel/cli/internal/ui"
	"github.com/urfave/cli/v3"
)

func LoginCommand() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Authenticate with Bluesky",
		Description: `Authenticate with Bluesky using one of two methods:

   1. Direct credentials via flags:
      skycli login --handle @user.bsky.social --password your-app-password

   2. Credentials from an env file:
      skycli login --file /path/to/.env

   The env file should contain:
      BLUESKY_HANDLE=your.handle.bsky.social
      BLUESKY_PASSWORD=your-app-password

   File paths can be relative or absolute.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path to env file containing BLUESKY_HANDLE and BLUESKY_PASSWORD",
			},
			&cli.StringFlag{
				Name:    "handle",
				Aliases: []string{"u"},
				Usage:   "Your Bluesky handle (e.g., @user.bsky.social)",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Your app password",
			},
		},
		Action: LoginAction,
	}
}

func LoginAction(ctx context.Context, cmd *cli.Command) error {
	if err := setup.EnsurePersistenceReady(ctx); err != nil {
		return fmt.Errorf("persistence layer not ready: %w", err)
	}

	reg := registry.Get()

	var handle, password string
	filePath := cmd.String("file")

	if filePath != "" {
		env, err := imports.ParseEnvFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse env file: %w", err)
		}

		handle = env["BLUESKY_HANDLE"]
		password = env["BLUESKY_PASSWORD"]

		if handle == "" {
			return fmt.Errorf("BLUESKY_HANDLE not found in env file")
		}
		if password == "" {
			return fmt.Errorf("BLUESKY_PASSWORD not found in env file")
		}
	} else {
		handle = cmd.String("handle")
		password = cmd.String("password")

		if handle == "" || password == "" {
			return fmt.Errorf("either --file or both --handle and --password are required")
		}
	}

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

	session, err := createSessionFromService(service, handle)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if err := sessionRepo.Save(ctx, session); err != nil {
		logger.Error("Failed to save session", "error", err)
		return fmt.Errorf("authentication succeeded but failed to save session: %w", err)
	}

	logger.Debug("Session saved successfully", "did", session.ID(), "handle", handle)
	ui.Successln("Successfully authenticated as %s", handle)
	return nil
}

// createSessionFromService creates a SessionModel from an authenticated service
func createSessionFromService(service *store.BlueskyService, handle string) (*store.SessionModel, error) {
	did := service.GetDid()
	if did == "" {
		return nil, fmt.Errorf("no DID available from authenticated service")
	}

	accessToken := service.GetAccessToken()
	refreshToken := service.GetRefreshToken()

	session := &store.SessionModel{
		Handle:     handle,
		Token:      accessToken + "|" + refreshToken,
		ServiceURL: service.BaseURL(),
		IsValid:    true,
	}
	session.SetID(did)

	return session, nil
}
