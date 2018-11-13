package registryfrontend

import (
	"context"
	"time"
)

type Registry struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

type TagInfo struct {
	Created       time.Time
	DockerVersion string
	EntryPoint    []string
	ExposedPorts  []string
	Layers        int
	Size          int64
	User          string
	Volumes       []string
}

type Client interface {
	Name() string
	URL() string

	Repositories(ctx context.Context) ([]string, error)
	RepositoriesN(ctx context.Context, n int, last string) ([]string, error)

	Tags(ctx context.Context, repository string) ([]string, error)
	TagsN(ctx context.Context, repository string, n int, last string) ([]string, error)

	Tag(ctx context.Context, repository, tag string) (*TagInfo, error)
}

type Storage interface {
	Registries() ([]Client, error)
	Registry(name string) (Client, error)
	Add(r Registry) error
	Update(r Registry) error
	Clear() error
	Remove(r Registry) error
}
