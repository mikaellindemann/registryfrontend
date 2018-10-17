package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mikaellindemann/registryfrontend"
	"github.com/mikaellindemann/registryfrontend/http"
	"github.com/mikaellindemann/templateloader"

	"github.com/sirupsen/logrus"
)

func main() {
	stop := make(chan os.Signal)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log := &logrus.Logger{
		Level: logrus.DebugLevel,
		Out:   os.Stderr,
		Hooks: make(logrus.LevelHooks),
		Formatter: &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "@timestamp",
				logrus.FieldKeyMsg:  "message",
			},
		},
	}

	storage := registryfrontend.NewInMemoryStorage()

	name := os.Getenv("REGISTRY_NAME")
	url := os.Getenv("REGISTRY_URL")
	user := os.Getenv("REGISTRY_AUTH_BASIC_USER")
	password := os.Getenv("REGISTRY_AUTH_BASIC_PASSWORD")

	if name != "" && url != "" {
		storage.Add(registryfrontend.Registry{
			Name:     name,
			Url:      url,
			User:     user,
			Password: password,
		})
	}

	var t templateloader.Loader

	if strings.ToUpper(os.Getenv("ENVIRONMENT")) == "DEVELOPMENT" {
		t = templateloader.NewOnRequestLoader()
		log.Debugln("Using the on-request-loader")
	} else {
		t = templateloader.NewPreloader()
		log.Debugln("Preloading templates")
	}

	s := http.NewServer(log, t, storage)
	s.Start()

	<-stop

	if err := s.Shutdown(); err != nil {
		panic(err)
	}
}
