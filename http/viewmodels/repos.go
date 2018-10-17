package viewmodels

type Repository struct {
	Name         string
	NumberOfTags int
}

type RegistryDetail struct {
	Title        string
	Registry     string
	Repositories []Repository
}
