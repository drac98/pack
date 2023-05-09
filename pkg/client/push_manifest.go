package client

import (
	"context"

	"github.com/buildpacks/imgutil/local"
	"github.com/buildpacks/imgutil/remote"
)

type PushManifestOptions struct {
	Index string
	Path  string
}

func (c *Client) PushManifest(ctx context.Context, opts PushManifestOptions) error {
	indexManifest, err := local.GetIndexManifest(opts.Index, opts.Path)
	if err != nil {
		panic(err)
	}

	idx, err := remote.NewIndex(opts.Index, c.keychain, remote.WithManifest(indexManifest))
	if err != nil {
		panic(err)
	}

	// Store index
	err = idx.Save()
	if err != nil {
		panic(err)
	}

	return nil
}
