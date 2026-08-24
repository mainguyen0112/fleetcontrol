package satellite

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

func TestToDomainCreateMapsRequestAndLeavesServiceDefaultsUnset(t *testing.T) {
	got := ToDomainCreate(gen.CreateSatelliteRequest{Name: "edge-01", Region: "ap-southeast"})
	if got.Name != "edge-01" || got.Region != "ap-southeast" {
		t.Fatalf("request fields were not mapped: %+v", got)
	}
	if got.Status != "" || got.ManagedBy != "" {
		t.Fatalf("mapper must leave service-owned defaults unset: %+v", got)
	}
}

func TestToResponseMapsAllFields(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2026, time.August, 24, 1, 2, 3, 0, time.UTC)
	lastSeenAt := createdAt.Add(time.Minute)
	sat := &Satellite{
		ID: id, Name: "edge-01", Region: "ap-southeast", Status: "Ready",
		ManagedBy: "operator", LastSeenAt: &lastSeenAt, CreatedAt: createdAt,
	}

	got := ToResponse(sat)
	if got.Id == nil || *got.Id != id || got.Name == nil || *got.Name != sat.Name ||
		got.Region == nil || *got.Region != sat.Region || got.Status == nil || string(*got.Status) != sat.Status ||
		got.ManagedBy == nil || string(*got.ManagedBy) != sat.ManagedBy || got.CreatedAt == nil || !got.CreatedAt.Equal(createdAt) ||
		got.LastSeenAt == nil || !got.LastSeenAt.Equal(lastSeenAt) {
		t.Fatalf("domain satellite was not fully mapped: %+v", got)
	}
}

func TestToResponseNilReturnsEmptyModel(t *testing.T) {
	if got := ToResponse(nil); got != (gen.Satellite{}) {
		t.Fatalf("expected empty response for nil satellite, got %+v", got)
	}
}
