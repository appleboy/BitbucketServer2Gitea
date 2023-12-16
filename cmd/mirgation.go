package cmd

import (
	"context"
)

type migration struct {
	ctx       context.Context
	bitbucket *bitbucket
	gitea     *gitea
}

func NewMigration(ctx context.Context) (*migration, error) {
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
	}

	return m, nil
}
