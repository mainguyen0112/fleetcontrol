package main

import (
	"net/http"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
	"github.com/mainguyen0112/fleetcontrol/api/internal/auth"
	"github.com/mainguyen0112/fleetcontrol/api/internal/health"
	"github.com/mainguyen0112/fleetcontrol/api/internal/satellite"
	"github.com/mainguyen0112/fleetcontrol/api/internal/user"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Server adapts the domain-idiomatic handlers (satellite.Handler, user.Handler,
// auth.Handler, health.Handler) to gen.ServerInterface. It exists purely as a
// naming/shape translation layer between OpenAPI-generated method names and
// the handlers' own method names — no business logic lives here.
type Server struct {
	sat  *satellite.Handler
	usr  *user.Handler
	auth *auth.Handler
	hlth *health.Handler
}

func NewServer(sat *satellite.Handler, usr *user.Handler, authH *auth.Handler, hlth *health.Handler) *Server {
	return &Server{sat: sat, usr: usr, auth: authH, hlth: hlth}
}

var _ gen.ServerInterface = (*Server)(nil)

// --- auth ---

func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	s.auth.Login(w, r)
}

// --- health ---

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	s.hlth.Health(w, r)
}

func (s *Server) GetVersion(w http.ResponseWriter, r *http.Request) {
	s.hlth.Version(w, r)
}

// --- satellites ---
//
// NOTE: ServerInterfaceWrapper already parses {id} from the URL and passes
// it here as a typed openapi_types.UUID. The underlying satellite.Handler
// methods still re-parse it themselves via chi.URLParam(r, "id") — this
// still works because both are registered on the same chi router, so the
// URL param is present in the request context either way. The `id`
// argument below is therefore currently unused; this is a deliberate,
// documented trade-off to avoid touching satellite/handler.go again in
// this step.

func (s *Server) GetSatellites(w http.ResponseWriter, r *http.Request) {
	s.sat.List(w, r)
}

func (s *Server) PostSatellites(w http.ResponseWriter, r *http.Request) {
	s.sat.Create(w, r)
}

func (s *Server) GetSatellitesId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	s.sat.GetByID(w, r)
}

func (s *Server) PatchSatellitesId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	s.sat.Update(w, r)
}

func (s *Server) DeleteSatellitesId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	s.sat.Delete(w, r)
}

func (s *Server) PostSatellitesIdHeartbeat(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	s.sat.Heartbeat(w, r)
}

// --- users ---

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	s.usr.List(w, r)
}

func (s *Server) PostUsers(w http.ResponseWriter, r *http.Request) {
	s.usr.Create(w, r)
}

func (s *Server) DeleteUsersId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	s.usr.Delete(w, r)
}
