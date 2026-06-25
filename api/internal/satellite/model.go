package satellite

import (
	"time"

	"github.com/google/uuid"
)

type Satellite struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Region     string     `json:"region"`
	Status     string     `json:"status"`
	ManagedBy  string     `json:"managed_by"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
