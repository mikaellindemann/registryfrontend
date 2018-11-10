package registryfrontend

import "context"

type Client interface {
	Name() string
	URL() string

	Repositories(ctx context.Context) ([]string, error)
	RepositoriesN(ctx context.Context, n int, last string) ([]string, error)

	Tags(ctx context.Context, repository string) ([]string, error)
	TagsN(ctx context.Context, repository string, n int, last string) ([]string, error)

	Tag(ctx context.Context, repository, tag string) (*TagInfo, error)
}
