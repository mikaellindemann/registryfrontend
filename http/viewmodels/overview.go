package viewmodels

type Registry struct {
	Name          string
	URL           string
	Online        bool
	NumberOfRepos int
}

type Overview struct {
	Title            string
	Registries       []Registry
	AddRemoveEnabled bool
}
