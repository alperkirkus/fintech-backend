package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID       `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id"`
	Action     string          `json:"action"`
	Details    json.RawMessage `json:"details,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
