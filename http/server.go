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

type Server interface {
	Start()
	Shutdown() error
}

type server struct {
	s http.Server
	l *logrus.Logger
}

func (s *server) Start() {
	go func() {
		err := s.s.ListenAndServe()

		if err != http.ErrServerClosed {
			panic(err)
		}
	}()
	s.l.WithField("address", s.s.Addr).Infof("Now listenening on %s.", s.s.Addr)
}

func (s *server) Shutdown() error {
	t := time.Now()
	defer func() {
		e := time.Since(t)
		s.l.WithField("duration", e.Nanoseconds()).Debugf("Shutdown took %s", e.String())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.s.Shutdown(ctx)
}

func must(h http.HandlerFunc, err error) http.HandlerFunc {
	if err != nil {
		panic(err)
	}
	return h
}

func NewServer(l *logrus.Logger, t templateloader.Loader, s registryfrontend.Storage) Server {
	router := mux.NewRouter()

	router.HandleFunc("/", must(overview(t, s))).Methods(http.MethodGet)
	router.HandleFunc("/test_connection", testConnection()).Methods(http.MethodPost)

	router.HandleFunc("/add_registry", addRegistryGet()).Methods(http.MethodGet)
	router.HandleFunc("/add_registry", addRegistryPost(s)).Methods(http.MethodPost)

	//router.HandleFunc("/update_registry", updateRegistryGet(s)).Methods(http.MethodGet)
	//router.HandleFunc("/update_registry", updateRegistryPost(s)).Methods(http.MethodPost)

	router.HandleFunc("/remove_registry", removeRegistry(s)).Methods(http.MethodPost)

	//router.HandleFunc("/delete_repo", deleteRepo(s)).Methods(http.MethodPost)

	//router.HandleFunc("/delete_tag", deleteTag(s)).Methods(http.MethodPost)

	router.HandleFunc("/registry/{registry}", must(repoOverview(t, s))).Methods(http.MethodGet)

	router.HandleFunc("/registry/{registry}/{repo}", must(tagOverview(t, s))).Methods(http.MethodGet)

	router.HandleFunc("/registry/{registry}/{repo}/{tag}", must(tagDetail(t, s))).Methods(http.MethodGet)

	return &server{
		s: http.Server{
			Addr:    ":8080",
			Handler: router,
		},
		l: l,
	}
}

func overview(tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
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
				// TODO: Log this error
				fmt.Println(err)
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

func addRegistryGet() http.HandlerFunc {
	// GET only
	return nil
}

func addRegistryPost(s registryfrontend.Storage) http.HandlerFunc {
	// POST only
	return nil
}

//func updateRegistryGet(s registryfrontend.Storage) http.HandlerFunc {
//	// GET Only
//	return nil
//}
//
//func updateRegistryPost(s registryfrontend.Storage) http.HandlerFunc {
//	// POST Only
//	return nil
//}

func removeRegistry(s registryfrontend.Storage) http.HandlerFunc {
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

		err = s.Remove(reg)

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

//func deleteRepo(s registryfrontend.Storage) http.HandlerFunc {
//	// POST only
//	return nil
//}
//
//func deleteTag(s registryfrontend.Storage) http.HandlerFunc {
//	// POST only
//	return nil
//}

func repoOverview(tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
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
				// TODO: Log this error
				fmt.Println(err)
			}
		},
		"http/templates/repos.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-repos.tmpl",
	)
}

func tagOverview(tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
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
				// TODO: Log this error
				fmt.Println(err)
			}
		},
		"http/templates/tags.tmpl", "http/templates/layout.tmpl", "http/templates/menu/menu-tags.tmpl",
	)
}

func tagDetail(tl templateloader.Loader, s registryfrontend.Storage) (http.HandlerFunc, error) {
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

			err = t.Execute(w, struct {
				Title         string
				Registry      string
				Repository    string
				Tag           string
				Created       string
				DockerVersion string
				Size          string
				Layers        int
				User          string
				Ports         string
				Volumes       string
			}{
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
				// TODO: Log this error
				fmt.Println(err)
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
