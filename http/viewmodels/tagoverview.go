package viewmodels

type TagOverviewInfo struct {
	Name string
	Created string
	Size string
	Layers int
}

type TagOverview struct {
	Title      string
	Registry   string
	Repository string
	Tags       []TagOverviewInfo
}
