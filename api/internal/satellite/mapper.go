package satellite

import (
	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

// ToDomainCreate converts an OpenAPI create request into the domain model.
//
// Business fields (ID, Status, ManagedBy, CreatedAt, LastSeenAt)
// are intentionally left zero-valued.
// Service.Create is responsible for initializing them.
func ToDomainCreate(req gen.CreateSatelliteRequest) Satellite {
	return Satellite{
		Name:   req.Name,
		Region: req.Region,
	}
}

// ToResponse converts the domain Satellite into the generated API model.
func ToResponse(s *Satellite) gen.Satellite {
	if s == nil {
		return gen.Satellite{}
	}

	id := s.ID
	name := s.Name
	region := s.Region
	status := gen.SatelliteStatus(s.Status)
	managedBy := gen.SatelliteManagedBy(s.ManagedBy)
	createdAt := s.CreatedAt

	return gen.Satellite{
		Id:         &id,
		Name:       &name,
		Region:     &region,
		Status:     &status,
		ManagedBy:  &managedBy,
		LastSeenAt: s.LastSeenAt,
		CreatedAt:  &createdAt,
	}
}

// ToResponseList converts a slice of domain Satellites into generated API models.
func ToResponseList(satellites []*Satellite) []gen.Satellite {
	resp := make([]gen.Satellite, 0, len(satellites))

	for _, s := range satellites {
		resp = append(resp, ToResponse(s))
	}

	return resp
}
