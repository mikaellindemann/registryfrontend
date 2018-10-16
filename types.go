package registry_frontend

type Registry struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}