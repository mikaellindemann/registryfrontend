package viewmodels

type Repository struct {
	Name         string
	UrlName      string
	NumberOfTags int
}

type RegistryDetail struct {
	Title        string
	Registry     string
	Repositories []Repository
}
