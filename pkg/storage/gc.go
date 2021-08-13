package storage

import (
	"context"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/storage"
	"github.com/distribution/distribution/v3/registry/storage/driver"
	"github.com/distribution/distribution/v3/registry/storage/driver/factory"
	"github.com/docker/libtrust"
	"github.com/octohelm/registry-proxy-cache/pkg/container"
	pkgerrors "github.com/pkg/errors"
)

func ClusterGarbageCollect(ctx context.Context, m map[string]*configuration.Configuration, hubGroupedContainerImages map[string]map[string]*container.ContainerImage, dryRun bool) error {
	for hub := range m {
		dc := m[hub]

		if dc.Storage == nil {
			continue
		}

		sr, d, err := NewStorageRegistryAndDriver(ctx, dc.Storage)
		if err != nil {
			return err
		}

		if err := UntagUnused(ctx, sr, hubGroupedContainerImages[hub]); err != nil {
			return pkgerrors.Wrapf(err, "[%s] untag unused failed:", hub)
		}

		if err := storage.MarkAndSweep(ctx, d, sr, storage.GCOpts{
			DryRun:         dryRun,
			RemoveUntagged: true,
		}); err != nil {
			return pkgerrors.Wrapf(err, "[%s] sweep failed:", hub)
		}
	}

	return nil
}

func NewStorageRegistryAndDriver(ctx context.Context, cstorage configuration.Storage) (distribution.Namespace, driver.StorageDriver, error) {
	d, err := factory.Create(cstorage.Type(), cstorage.Parameters())
	if err != nil {
		return nil, nil, err
	}

	k, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		return nil, nil, err
	}

	sr, err := storage.NewRegistry(ctx, d, storage.Schema1SigningKey(k))
	if err != nil {
		return nil, nil, err
	}
	return sr, d, err
}

func UntagUnused(ctx context.Context, sr distribution.Namespace, used map[string]*container.ContainerImage) error {
	re := sr.(distribution.RepositoryEnumerator)

	return re.Enumerate(ctx, func(repoName string) error {
		named, err := reference.WithName(repoName)
		if err != nil {
			return pkgerrors.Wrapf(err, "failed to parse repo name %s", repoName)
		}
		repository, err := sr.Repository(ctx, named)
		if err != nil {
			return pkgerrors.Wrapf(err, "failed to construct repository")
		}

		tagService := repository.Tags(ctx)

		tags, err := tagService.All(ctx)
		if err != nil {
			return pkgerrors.Wrapf(err, "failed to list all tags")
		}

		for _, t := range tags {
			fullRef := named.String() + ":" + t

			if _, ok := used[fullRef]; ok {
				continue
			}

			if err := tagService.Untag(ctx, t); err != nil {
				return pkgerrors.Wrapf(err, "untag %s", fullRef)
			}
		}

		return nil
	})
}
