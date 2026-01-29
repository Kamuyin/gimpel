package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/master/api"
)

func (s *Server) RegisterRESTAPIs(mux *http.ServeMux) {
	moduleAPI := api.NewModuleAPI(s.Store, s.Verifier, s.ModuleKey)
	deploymentAPI := api.NewDeploymentAPI(s.Store)
	pairingAPI := api.NewPairingAPI(s.Store)

	corsMiddleware := func(h http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			h(w, r)
		})
	}

	mux.Handle("POST /api/v1/modules", corsMiddleware(moduleAPI.HandleUploadModule))
	mux.Handle("GET /api/v1/modules", corsMiddleware(moduleAPI.HandleListModules))
	mux.Handle("GET /api/v1/modules/{id}/{version}", corsMiddleware(moduleAPI.HandleGetModule))
	mux.Handle("GET /api/v1/modules/{id}/{version}/download", corsMiddleware(moduleAPI.HandleDownloadModule))
	mux.Handle("DELETE /api/v1/modules/{id}/{version}", corsMiddleware(moduleAPI.HandleDeleteModule))

	mux.Handle("POST /api/v1/satellites/{id}/deployments", corsMiddleware(deploymentAPI.HandleCreateDeployment))
	mux.Handle("GET /api/v1/satellites/{id}/deployments", corsMiddleware(deploymentAPI.HandleGetDeployment))
	mux.Handle("DELETE /api/v1/satellites/{id}/deployments", corsMiddleware(deploymentAPI.HandleDeleteDeployment))

	mux.Handle("GET /api/v1/satellites", corsMiddleware(deploymentAPI.HandleListSatellites))
	mux.Handle("GET /api/v1/satellites/{id}", corsMiddleware(deploymentAPI.HandleGetSatellite))

	mux.Handle("GET /api/v1/deployments", corsMiddleware(deploymentAPI.HandleListDeployments))

	mux.Handle("POST /api/v1/pairings", corsMiddleware(pairingAPI.HandleCreatePairing))

	mux.Handle("GET /api/v1/module-signing-key", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if s.ModuleKeyPrivateServed || len(s.ModuleKeyPrivatePEM) == 0 {
			http.Error(w, "not available", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", "attachment; filename=module-signing.key")
		w.Write(s.ModuleKeyPrivatePEM)
		s.ModuleKeyPrivatePEM = nil
		s.ModuleKeyPrivateServed = true
	}))

	log.Info("REST API handlers registered")
}

func (s *Server) StartRESTServer(address string) error {
	mux := http.NewServeMux()

	s.RegisterRESTAPIs(mux)

	//mux.Handle("GET /{path...}", webUIHandler())

	log.WithField("address", address).Info("REST API server starting")

	go func() {
		if err := http.ListenAndServe(address, mux); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("REST API server error")
		}
	}()

	return nil
}
