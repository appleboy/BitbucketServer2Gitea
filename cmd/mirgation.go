package cmd

import (
	"context"
	"log/slog"
	"os"
)

type migration struct {
	ctx       context.Context
	bitbucket *bitbucket
	gitea     *gitea
	logger    *slog.Logger
}

func NewMigration(ctx context.Context) (*migration, error) {
	logLevel := &slog.LevelVar{} // INFO
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)

	if debug {
		logLevel.Set(slog.LevelDebug)
	}

	// initial bitbucket client
	b, err := NewBitbucket(ctx)
	if err != nil {
		return nil, err
	}

	g, err := NewGitea(ctx)
	if err != nil {
		return nil, err
	}

	m := &migration{
		ctx:       ctx,
		bitbucket: b,
		gitea:     g,
		logger:    slog.New(handler),
	}

	return m, nil
}
