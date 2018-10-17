package registry_frontend

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/registry/client"
	"github.com/pkg/errors"
)

func (r Registry) Repositories(ctx context.Context) ([]string, error) {
	reg, err := client.NewRegistry(r.Url, NewBasicAuthRoundTripper(r.Url, r.User, r.Password))

	if err != nil {
		return nil, errors.Wrap(err, "failed creating registry client")
	}

	repos := make([]string, 10)
	res := make([]string, 0, 10)
	last := ""

	for err != io.EOF {
		var n int
		n, err = reg.Repositories(ctx, repos, last)

		if err != nil && err != io.EOF {
			return nil, errors.Wrap(err, "failed fetching repositories")
		}

		res = append(res, repos[:n]...)

		last = repos[n-1]
	}

	return res, nil
}

type named string

func (n named) Name() string {
	return string(n)
}

func (n named) String() string {
	return string(n)
}

func (r Registry) Tags(ctx context.Context, repo string) ([]string, error) {
	repoService, err := client.NewRepository(named(repo), r.Url, NewBasicAuthRoundTripper(r.Url, r.User, r.Password))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository client")
	}

	tags, err := repoService.Tags(ctx).All(ctx)

	return tags, errors.Wrap(err, "failed to fetch tags")
}

func (r Registry) Tag(ctx context.Context, repo, tag string) (*TagInfo, error) {
	repoService, err := client.NewRepository(named(repo), r.Url, NewBasicAuthRoundTripper(r.Url, r.User, r.Password))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository client")
	}

	desc, err := repoService.Tags(ctx).Get(ctx, tag)

	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch tag")
	}

	ms, err := repoService.Manifests(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create manifest service")
	}

	m, err := ms.Get(ctx, desc.Digest, distribution.WithTag(tag), distribution.WithManifestMediaTypes([]string{desc.MediaType}))

	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch manifest")
	}

	manifest, ok := m.(*schema1.SignedManifest)

	if !ok {
		return nil, errors.Wrap(err, "unexpected manifest schema version")
	}

	ti := tagInfo{}
	err = json.Unmarshal([]byte(manifest.History[0].V1Compatibility), &ti)

	if err != nil {
		return nil, errors.Wrap(err, "unable to parse tag information")
	}

	size := int64(0)
	for _, layer := range manifest.FSLayers {
		ls, err := CacheOrGetLayerSize(ctx, repoService, layer.BlobSum)

		if err != nil {
			return nil, errors.Wrap(err, "failed fetching blob")
		}

		size += ls
	}

	res := &TagInfo{
		Created:       ti.Created,
		ExposedPorts:  keys(ti.Config.ExposedPorts),
		Volumes:       keys(ti.Config.Volumes),
		EntryPoint:    ti.Config.EntryPoint,
		User:          ti.Config.User,
		DockerVersion: ti.DockerVersion,
		Layers:        len(manifest.FSLayers),
		Size:          size,
	}

	return res, errors.Wrap(err, "failed to fetch tag")
}

func keys(m map[string]interface{}) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

type config struct {
	ExposedPorts map[string]interface{}
	Volumes      map[string]interface{}
	EntryPoint   []string
	User         string
}

type tagInfo struct {
	Created       time.Time `json:"created"`
	Config        config    `json:"config"`
	DockerVersion string    `json:"docker_version"`
}
