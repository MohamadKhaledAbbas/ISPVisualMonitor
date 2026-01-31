package dto

import (
	"time"

	"github.com/google/uuid"
)

// TenantDTO represents a tenant in API responses
type TenantDTO struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Slug             string     `json:"slug"`
	ContactEmail     string     `json:"contact_email"`
	ContactPhone     *string    `json:"contact_phone,omitempty"`
	SubscriptionTier string     `json:"subscription_tier"`
	MaxDevices       int        `json:"max_devices"`
	MaxUsers         int        `json:"max_users"`
	Status           string     `json:"status"`
	TrialEndsAt      *time.Time `json:"trial_ends_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	Name             string  `json:"name" validate:"required"`
	Slug             string  `json:"slug" validate:"required"`
	ContactEmail     string  `json:"contact_email" validate:"required,email"`
	ContactPhone     *string `json:"contact_phone,omitempty"`
	SubscriptionTier string  `json:"subscription_tier" validate:"required"`
	MaxDevices       int     `json:"max_devices" validate:"required,min=1"`
	MaxUsers         int     `json:"max_users" validate:"required,min=1"`
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name             *string `json:"name,omitempty"`
	ContactEmail     *string `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone     *string `json:"contact_phone,omitempty"`
	SubscriptionTier *string `json:"subscription_tier,omitempty"`
	MaxDevices       *int    `json:"max_devices,omitempty" validate:"omitempty,min=1"`
	MaxUsers         *int    `json:"max_users,omitempty" validate:"omitempty,min=1"`
	Status           *string `json:"status,omitempty"`
}
