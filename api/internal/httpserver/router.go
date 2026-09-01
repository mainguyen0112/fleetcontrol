package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mainguyen0112/fleetcontrol/api/gen"
	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/docs"
)

type Middleware func(http.Handler) http.Handler

func NewRouter(server gen.ServerInterface, docsHandler *docs.Handler, jwtSecret string, requestLogger Middleware) http.Handler {
	wrapper := &gen.ServerInterfaceWrapper{Handler: server}
	r := chi.NewRouter()
	if requestLogger != nil {
		r.Use(requestLogger)
	}

	r.Post("/auth/login", wrapper.PostAuthLogin)
	r.Get("/health", wrapper.GetHealth)
	r.Get("/version", wrapper.GetVersion)
	r.Get("/openapi.yaml", docsHandler.ServeSpec)
	r.Get("/docs", docsHandler.ServeUI)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Post("/satellites", wrapper.PostSatellites)
		r.Get("/satellites", wrapper.GetSatellites)
		r.Get("/satellites/{id}", wrapper.GetSatellitesId)
		r.Patch("/satellites/{id}", wrapper.PatchSatellitesId)
		r.Delete("/satellites/{id}", wrapper.DeleteSatellitesId)
		r.Post("/satellites/{id}/heartbeat", wrapper.PostSatellitesIdHeartbeat)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Use(auth.RequireHumanRole(auth.RoleAdmin))
		r.Post("/users", wrapper.PostUsers)
		r.Get("/users", wrapper.GetUsers)
		r.Delete("/users/{id}", wrapper.DeleteUsersId)
	})

	return r
}
