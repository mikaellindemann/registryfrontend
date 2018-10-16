package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"registry-frontend"
	"time"
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

func NewServer(l *logrus.Logger, s registry_frontend.Storage) Server {
	router := mux.NewRouter()

	router.HandleFunc("/", overview(s)).Methods(http.MethodGet)
	router.HandleFunc("/test_connection", testConnection()).Methods(http.MethodPost)

	router.HandleFunc("/add_registry", addRegistryGet()).Methods(http.MethodGet)
	router.HandleFunc("/add_registry", addRegistryPost(s)).Methods(http.MethodPost)

	//router.HandleFunc("/update_registry", updateRegistryGet(s)).Methods(http.MethodGet)
	//router.HandleFunc("/update_registry", updateRegistryPost(s)).Methods(http.MethodPost)

	router.HandleFunc("/remove_registry", removeRegistry(s)).Methods(http.MethodPost)

	//router.HandleFunc("/delete_repo", deleteRepo(s)).Methods(http.MethodPost)

	//router.HandleFunc("/delete_tag", deleteTag(s)).Methods(http.MethodPost)

	router.HandleFunc("/registry/{registry}", repoOverview(s)).Methods(http.MethodGet)

	router.HandleFunc("/registry/{registry}/{repo}", tagOverview(s)).Methods(http.MethodGet)

	router.HandleFunc("/registry/{registry}/{repo}/{tag}", tagDetail(s)).Methods(http.MethodGet)

	return &server{
		s: http.Server{
			Addr:    ":8080",
			Handler: router,
		},
		l: l,
	}
}

func overview(s registry_frontend.Storage) http.HandlerFunc {
	t := template.Must(template.New("registries").ParseFiles("http/templates/registries.tmpl", "http/templates/header.tmpl", "http/templates/footer.tmpl"))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		rs, err := s.Registries()

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusInternalServerError)).Error(), http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, struct {
			Title      string
			Registries []registry_frontend.Registry
		}{Title: "Registries", Registries: rs})

		if err != nil {
			// TODO: Log this error
			fmt.Println(err)
		}
	}
}

func testConnection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reg := registry_frontend.Registry{}

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

func addRegistryPost(s registry_frontend.Storage) http.HandlerFunc {
	// POST only
	return nil
}

//func updateRegistryGet(s registry_frontend.Storage) http.HandlerFunc {
//	// GET Only
//	return nil
//}
//
//func updateRegistryPost(s registry_frontend.Storage) http.HandlerFunc {
//	// POST Only
//	return nil
//}

func removeRegistry(s registry_frontend.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reg := registry_frontend.Registry{}

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

//func deleteRepo(s registry_frontend.Storage) http.HandlerFunc {
//	// POST only
//	return nil
//}
//
//func deleteTag(s registry_frontend.Storage) http.HandlerFunc {
//	// POST only
//	return nil
//}

func repoOverview(s registry_frontend.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("repos").ParseFiles("http/templates/repos.tmpl", "http/templates/header.tmpl", "http/templates/footer.tmpl"))
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

		err = t.Execute(w, struct {
			Title string
			Registry string
			Repositories []string
		}{
			Title: "Repositories",
			Registry: reg.Name,
			Repositories: repos,
		})


		if err != nil {
			// TODO: Log this error
			fmt.Println(err)
		}
	}
}

func tagOverview(s registry_frontend.Storage) http.HandlerFunc {
	t := template.Must(template.New("tags").ParseFiles("http/templates/tags.tmpl", "http/templates/header.tmpl", "http/templates/footer.tmpl"))
	return func(w http.ResponseWriter, r *http.Request) {
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

		tags, err := reg.Tags(r.Context(), vars["repo"])

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
			return
		}

		err = t.Execute(w, struct {
			Title string
			Repository string
			Tags []string
		}{
			Title: "Tags",
			Repository: vars["repo"],
			Tags: tags,
		})

		if err != nil {
			// TODO: Log this error
			fmt.Println(err)
		}
	}
}

func tagDetail(s registry_frontend.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		t, err := reg.Tag(r.Context(), vars["repo"], vars["tag"])

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusNotFound)).Error(), http.StatusNotFound)
			return
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		err = enc.Encode(t)

		if err != nil {
			http.Error(w, errors.Wrap(err, http.StatusText(http.StatusInternalServerError)).Error(), http.StatusInternalServerError)
		}
	}
}