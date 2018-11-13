package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mikaellindemann/registryfrontend"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type V2Client struct {
	name string
	url  string
	c    http.Client
}

func MakeV2(name, baseUri string) (*V2Client, error) {
	u, err := url.Parse(baseUri)

	if err != nil {
		return nil, err
	}

	return newV2(name, baseUri, &baseUrlRoundTripper{u.Scheme, u.Host, http.DefaultTransport}), nil
}

func MakeV2BasicAuth(name, baseUri, user, password string) (*V2Client, error) {
	u, err := url.Parse(baseUri)

	if err != nil {
		return nil, err
	}

	return newV2(name, baseUri, &baseUrlRoundTripper{
		u.Scheme,
		u.Host,
		&basicAuthRoundTripper{baseUri, user, password, http.DefaultTransport},
	}), nil
}

func newV2(name, url string, tripper http.RoundTripper) *V2Client {
	return &V2Client{
		name: name,
		url:  url,
		c: http.Client{
			Transport: tripper,
		},
	}
}

func (v *V2Client) Name() string {
	return v.name
}

func (v *V2Client) URL() string {
	return v.url
}

func (v *V2Client) Repositories(ctx context.Context) ([]string, error) {
	return v.RepositoriesN(ctx, -1, "")
}

type repositoriesDto struct {
	Repositories []string `json:"repositories"`
}

func (v *V2Client) RepositoriesN(ctx context.Context, n int, last string) ([]string, error) {
	var u string
	if n <= 0 {
		u = "/v2/_catalog"
	} else if last == "" {
		u = fmt.Sprintf("/v2/_catalog?n=%d", n)
	} else {
		u = fmt.Sprintf("/v2/_catalog?n=%d&last=%s", n, last)
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry request")
	}

	req = req.WithContext(ctx)

	resp, err := v.c.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "failed fetching repositories")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "could not read registry response")
	}

	dto := repositoriesDto{}

	err = json.Unmarshal(content, &dto)

	if err != nil {
		return nil, errors.Wrap(err, "could not parse registry response")
	}

	return dto.Repositories, nil
}

type tagsDto struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (v *V2Client) Tags(ctx context.Context, repository string) ([]string, error) {
	return v.TagsN(ctx, repository, -1, "")
}

func (v *V2Client) TagsN(ctx context.Context, repository string, n int, last string) ([]string, error) {
	var u string
	if n <= 0 {
		u = fmt.Sprintf("/v2/%s/tags/list", repository)
	} else if last == "" {
		u = fmt.Sprintf("/v2/%s/tags/list?n=%d", repository, n)
	} else {
		u = fmt.Sprintf("/v2/%s/tags/list?n=%d&last=%s", repository, n, last)
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry request")
	}

	req = req.WithContext(ctx)
	resp, err := v.c.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "failed fetching tags")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "could not read registry response")
	}

	dto := tagsDto{}

	err = json.Unmarshal(content, &dto)

	if err != nil {
		return nil, errors.Wrap(err, "could not parse registry response")
	}

	return dto.Tags, nil
}

type fsLayer struct {
	BlobSum digest.Digest `json:"blobSum"`
}

type history struct {
	V1Compatibility string `json:"v1Compatibility"`
}

type manifestDto struct {
	FSLayers []fsLayer `json:"fsLayers"`
	History  []history `json:"history"`
}

type config struct {
	ExposedPorts map[string]interface{}
	Volumes      map[string]interface{}
	EntryPoint   []string
	User         string
}

type compatibilityInfo struct {
	Created       time.Time `json:"created"`
	Config        config    `json:"config"`
	DockerVersion string    `json:"docker_version"`
}

func (v *V2Client) Tag(ctx context.Context, repository, tag string) (*registryfrontend.TagInfo, error) {
	u := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)

	req, err := http.NewRequest(http.MethodGet, u, nil)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry request")
	}

	req = req.WithContext(ctx)
	resp, err := v.c.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "failed fetching tags")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "could not read registry response")
	}

	dto := manifestDto{}

	err = json.Unmarshal(content, &dto)

	if err != nil {
		return nil, errors.Wrap(err, "could not parse registry response")
	}

	info := compatibilityInfo{}

	err = json.Unmarshal([]byte(dto.History[0].V1Compatibility), &info)

	if err != nil {
		return nil, errors.Wrap(err, "could not parse tag information")
	}

	totalSize := int64(0)

	for _, l := range dto.FSLayers {
		s, err := v.BlobSize(ctx, repository, l.BlobSum)

		if err != nil {
			return nil, err
		}

		totalSize += s
	}

	var keys = func(m map[string]interface{}) []string {
		res := make([]string, 0, len(m))
		for k := range m {
			res = append(res, k)
		}
		return res
	}

	return &registryfrontend.TagInfo{
		Created:       info.Created,
		DockerVersion: info.DockerVersion,
		EntryPoint:    info.Config.EntryPoint,
		ExposedPorts:  keys(info.Config.ExposedPorts),
		Layers:        len(dto.FSLayers),
		Size:          totalSize,
		User:          info.Config.User,
		Volumes:       keys(info.Config.Volumes),
	}, nil
}

func (v *V2Client) BlobSize(ctx context.Context, repository string, d digest.Digest) (int64, error) {
	u := fmt.Sprintf("/v2/%s/blobs/%s", repository, d.String())

	req, err := http.NewRequest(http.MethodHead, u, nil)

	if err != nil {
		return 0, errors.Wrap(err, "failed to create registry request")
	}

	req = req.WithContext(ctx)
	resp, err := v.c.Do(req)

	if err != nil {
		return 0, errors.Wrap(err, "failed fetching tags")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	c := resp.Header.Get("content-length")

	s, err := strconv.ParseInt(c, 10, 64)

	return s, errors.Wrap(err, "failed to parse size of blob")
}
