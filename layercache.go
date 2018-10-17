package registryfrontend

import (
	"context"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

var cache = make(map[digest.Digest]int64)

func CacheOrGetLayerSize(ctx context.Context, rs distribution.Repository, d digest.Digest) (int64, error) {
	if val, ok := cache[d]; ok {
		return val, nil
	}

	blob, err := rs.Blobs(ctx).Stat(ctx, d)

	if err != nil {
		return -1, errors.Wrap(err, "failed fetching blob information")
	}

	cache[d] = blob.Size

	return blob.Size, nil
}
