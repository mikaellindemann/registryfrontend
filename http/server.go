package http

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/mikaellindemann/registryfrontend"
	"github.com/mikaellindemann/registryfrontend/http/viewmodels"
	"github.com/mikaellindemann/templateloader"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Server struct {
	s http.Server
	l *logrus.Logger
}

// Start makes the Server available.
// The server will run in a separate goroutine, and this function will return immediately.
func (s *Server) Start() {
	go func() {
		err := s.s.ListenAndServe()

		if err != http.ErrServerClosed {
			panic(err)
		}
	}()
	s.l.WithField("address", s.s.Addr).Infof("Now listenening on %s.", s.s.Addr)
}

// Shutdown will make the server unreachable.
// The goroutine created in Start, will have stopped running when this function returns.
func (s *Server) Shutdown() error {
	t := time.Now()
	defer func() {
		e := time.Since(t)
		s.l.WithField("duration", e.Nanoseconds()).Debugf("Shutdown took %s", e.String())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.s.Shutdown(ctx)
}

func NewServer(l *logrus.Logger, t templateloader.Loader, s registryfrontend.Storage) *Server {
	must := func(h http.HandlerFunc, err error) http.HandlerFunc {
		if err != nil {
			panic(err)
		}
		return h
	}

	router := mux.NewRouter()

	router.HandleFunc("/", must(overview(l, t, s))).Methods(http.MethodGet)
	router.HandleFunc("/test_connection", testConnection()).Methods(http.MethodPost)

	router.HandleFunc("/add_registry", must(addRegistryGet(l, t))).Methods(http.MethodGet)
	router.HandleFunc("/add_registry", addRegistryPost(s)).Methods(http.MethodPost)

	router.HandleFunc("/remove_registry", removeRegistry(s)).Methods(http.MethodPost)

	router.HandleFunc("/registry/{registry}", must(repoOverview(l, t, s))).Methods(http.MethodGet)

	router.HandleFunc("/registry/{registry}/{repo}", must(tagOverview(l, t, s))).Methods(http.MethodGet)

	router.HandleFunc("/registry/{registry}/{repo}/{tag}", must(tagDetail(l, t, s))).Methods(http.MethodGet)

	return &Server{
		s: http.Server{
			Addr:    ":8080",
			Handler: router,
		},
		l: l,
	}
}

func overview(l *logrus.Logger, tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
	return tl.Load(
		"layout",
		func(t *template.Template, w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			rs, err := s.Registries()

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusInternalServerError)).Error(), http.StatusInternalServerError)
				return
			}

			regs := make([]viewmodels.Registry, 0, len(rs))

			for _, reg := range rs {
				repos, err := reg.Repositories(r.Context())

				regs = append(regs, viewmodels.Registry{
					Name:          reg.Name,
					URL:           reg.Url,
					Online:        err == nil,
					NumberOfRepos: len(repos),
				})
			}

			err = t.Execute(w, viewmodels.Overview{
				Title:      "Registries",
				Registries: regs,
			})

			if err != nil {
				l.Errorf("%+v", err)
			}
		},
		"http/templates/registries.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-registries.tmpl",
	)
}

func testConnection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reg := registryfrontend.Registry{}

		err := json.NewDecoder(r.Body).Decode(&reg)

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusBadRequest)).Error(), http.StatusBadRequest)
			return
		}

		_, err = reg.Repositories(r.Context())

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusBadRequest)).Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func addRegistryGet(l *logrus.Logger, tl templateloader.Loader) (http.HandlerFunc, error) {
	return tl.Load(
		"layout",
		func(t *template.Template, w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			err := t.Execute(w, viewmodels.Layout{
				Title: "Add registry",
			})

			if err != nil {
				l.Errorf("%+v", err)
			}
		},
		"http/templates/registryform.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-registries.tmpl",
	)
}

func addRegistryPost(s registryfrontend.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reg := registryfrontend.Registry{}

		err := r.ParseForm()

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusBadRequest)).Error(), http.StatusBadRequest)
			return
		}

		reg.Name = r.Form.Get("name")
		reg.Url = r.Form.Get("url")
		reg.User = r.Form.Get("user")
		reg.Password = r.Form.Get("password")

		err = s.Add(reg)

		if err != nil {
			http.Error(w, fmt.Sprintf("%+v", errors.WithStack(errors.Wrap(err, http.StatusText(http.StatusInternalServerError)))), http.StatusInternalServerError)
			return
		}

		u := *r.URL
		u.Path = "/"

		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

func removeRegistry(s registryfrontend.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reg := registryfrontend.Registry{}

		err := r.ParseForm()

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusBadRequest)).Error(), http.StatusBadRequest)
			return
		}

		reg.Name = r.Form.Get("name")

		err = s.Remove(reg)

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func repoOverview(l *logrus.Logger, tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
	return tl.Load(
		"layout",
		func(t *template.Template, w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			vars := mux.Vars(r)

			reg, err := s.Registry(vars["registry"])

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
				return
			}

			repos, err := reg.Repositories(r.Context())

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusBadRequest)).Error(), http.StatusBadRequest)
				return
			}

			reps := make([]viewmodels.Repository, 0, len(repos))

			for _, repo := range repos {
				ti, err := reg.Tags(r.Context(), repo)

				if err != nil {
					http.Error(w, fmt.Sprintf("%+v", errors.WithStack(errors.Wrap(err, "failed fetching repository details"))), http.StatusInternalServerError)
					return
				}

				reps = append(reps, viewmodels.Repository{
					Name:         repo,
					NumberOfTags: len(ti),
				})
			}

			err = t.Execute(w, viewmodels.RegistryDetail{
				Title:        "Repositories",
				Registry:     reg.Name,
				Repositories: reps,
			})

			if err != nil {
				l.Errorf("%+v", err)
			}
		},
		"http/templates/repos.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-repos.tmpl",
	)
}

func tagOverview(l *logrus.Logger, tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
	return tl.Load(
		"layout",
		func(t *template.Template, w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			vars := mux.Vars(r)

			reg, err := s.Registry(vars["registry"])

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
				return
			}

			ts, err := reg.Tags(r.Context(), vars["repo"])

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
				return
			}

			tags := make([]viewmodels.TagOverviewInfo, 0, len(ts))

			for _, tag := range ts {
				ti, err := reg.Tag(r.Context(), vars["repo"], tag)

				if err != nil {
					http.Error(w, fmt.Sprintf("%+v", errors.WithStack(errors.Wrap(err, "failed fetching tag information"))), http.StatusInternalServerError)
					return
				}

				tags = append(tags, viewmodels.TagOverviewInfo{
					Name:    tag,
					Created: ti.Created.Format("January 2 2006 15:04:05"),
					Size:    sizeToString(ti.Size),
					Layers:  ti.Layers,
				})
			}

			sort.Slice(tags, func(i, j int) bool {
				// Errors ignored as the strings were created by applying this format.
				t1, _ := time.Parse("January 2 2006 15:04:05", tags[i].Created)
				t2, _ := time.Parse("January 2 2006 15:04:05", tags[j].Created)

				return t2.Before(t1)
			})

			err = t.Execute(w, viewmodels.TagOverview{
				Title:      "Tags",
				Registry:   vars["registry"],
				Repository: vars["repo"],
				Tags:       tags,
			})

			if err != nil {
				l.Errorf("%+v", err)
			}
		},
		"http/templates/tags.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-tags.tmpl",
	)
}

func tagDetail(l *logrus.Logger, tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
	return tl.Load(
		"layout",
		func(t *template.Template, w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			vars := mux.Vars(r)

			reg, err := s.Registry(vars["registry"])

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
				return
			}

			tag, err := reg.Tag(r.Context(), vars["repo"], vars["tag"])

			if err != nil {
				http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
				return
			}

			err = t.Execute(w, viewmodels.TagDetails{
				Title:         "Tag details",
				Registry:      vars["registry"],
				Repository:    vars["repo"],
				Tag:           vars["tag"],
				Created:       tag.Created.Format("January 2 2006 15:04:05 "),
				DockerVersion: tag.DockerVersion,
				Size:          sizeToString(tag.Size),
				Layers:        tag.Layers,
				User:          tag.User,
				Volumes:       fmt.Sprint(tag.Volumes),
				Ports:         fmt.Sprint(tag.ExposedPorts),
			})

			if err != nil {
				l.Errorf("%+v", err)
			}
		},
		"http/templates/tagdetails.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-tag-details.tmpl",
	)
}

func sizeToString(byteCount int64) string {

	if gb := float64(byteCount) / 1024.0 / 1024.0 / 1024.0; gb >= 1.0 {
		return fmt.Sprintf("%.2f GB", gb)
	} else if mb := float64(byteCount) / 1024.0 / 1024.0; mb >= 1.0 {
		return fmt.Sprintf("%.2f MB", mb)
	} else if kb := float64(byteCount) / 1024.0; kb >= 1.0 {
		return fmt.Sprintf("%.2f KB", kb)
	}
	return fmt.Sprintf("%d B", byteCount)
}
