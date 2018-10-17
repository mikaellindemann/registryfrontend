package registryfrontend

import "time"

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
