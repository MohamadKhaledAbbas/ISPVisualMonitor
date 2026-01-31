package dto

import (
	"time"

	"github.com/google/uuid"
)

// AlertDTO represents an alert in API responses
type AlertDTO struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	Severity        string     `json:"severity"`
	Status          string     `json:"status"`
	TargetType      *string    `json:"target_type,omitempty"`
	TargetID        *uuid.UUID `json:"target_id,omitempty"`
	TriggeredAt     time.Time  `json:"triggered_at"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy  *uuid.UUID `json:"acknowledged_by,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

// AcknowledgeAlertRequest represents the request to acknowledge an alert
type AcknowledgeAlertRequest struct {
	Note string `json:"note,omitempty"`
}
