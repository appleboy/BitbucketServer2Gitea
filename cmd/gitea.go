package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/spf13/viper"
)

// NewGitea creates a new instance of the gitea struct.
func NewGitea(ctx context.Context) (*gitea, error) {
	g := &gitea{
		ctx:        ctx,
		server:     viper.GetString("gitea.server"),
		token:      viper.GetString("gitea.token"),
		skipVerify: viper.GetBool("gitea.skip-verify"),
	}

	err := g.init()
	if err != nil {
		return nil, err
	}

	return g, nil
}

// gitea is a struct that holds the gitea client.
type gitea struct {
	ctx        context.Context
	server     string
	token      string
	skipVerify bool
	client     *gsdk.Client
}

// init initializes the gitea client.
func (g *gitea) init() error {
	if g.server == "" || g.token == "" {
		return errors.New("mission gitea server or token")
	}

	g.server = strings.TrimRight(g.server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.token),
	}

	if g.skipVerify {
		// add new http client for skip verify
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		opts = append(opts, gsdk.SetHTTPClient(httpClient))
	}

	client, err := gsdk.NewClient(g.server, opts...)
	if err != nil {
		return err
	}
	g.client = client

	return nil
}

// CreateOrgOption create organization option
type CreateOrgOption struct {
	Name        string
	Description string
	Visibility  bool
}

// CreateAndGetOrg create and get organization
func (g *gitea) CreateAndGetOrg(opts CreateOrgOption) (*gsdk.Organization, error) {
	newOrg, reponse, err := g.client.GetOrg(opts.Name)
	if reponse.StatusCode == http.StatusNotFound {
		visible := gsdk.VisibleTypePublic
		if !opts.Visibility {
			visible = gsdk.VisibleTypePrivate
		}
		newOrg, _, err = g.client.CreateOrg(gsdk.CreateOrgOption{
			Name:        opts.Name,
			Description: opts.Description,
			Visibility:  visible,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return newOrg, nil
}
