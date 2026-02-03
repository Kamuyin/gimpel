package server

import (
	"io/fs"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"gimpel/web"
)

func (s *Server) RegisterWebUI(mux *http.ServeMux) {
	distFS, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		log.WithError(err).Error("failed to create sub-filesystem for web UI")
		return
	}

	fileServer := http.FileServer(http.FS(distFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if !fs.ValidPath(path) {
			http.NotFound(w, r)
			return
		}

		f, err := distFS.Open(path)
		if err != nil {
			r.URL.Path = "/"
		} else {
			f.Close()
		}

		fileServer.ServeHTTP(w, r)
	})

	log.Info("WebUI handler registered")
}
