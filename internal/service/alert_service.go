package service

import (
	"context"
	"fmt"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/repository"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AlertService handles alert business logic
type AlertService struct {
	alertRepo repository.AlertRepository
	logger    *zap.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(
	alertRepo repository.AlertRepository,
	logger *zap.Logger,
) *AlertService {
	return &AlertService{
		alertRepo: alertRepo,
		logger:    logger,
	}
}

// GetAlert retrieves an alert by ID
func (s *AlertService) GetAlert(ctx context.Context, tenantID, alertID uuid.UUID) (dto.AlertDTO, error) {
	alert, err := s.alertRepo.GetByID(ctx, tenantID, alertID)
	if err != nil {
		s.logger.Error("Failed to get alert", zap.Error(err))
		return dto.AlertDTO{}, fmt.Errorf("alert not found")
	}

	return toAlertDTO(alert), nil
}

// ListAlerts retrieves a list of alerts
func (s *AlertService) ListAlerts(ctx context.Context, tenantID uuid.UUID, opts repository.ListOptions) ([]dto.AlertDTO, int64, error) {
	alerts, total, err := s.alertRepo.List(ctx, tenantID, opts)
	if err != nil {
		s.logger.Error("Failed to list alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list alerts")
	}

	alertDTOs := make([]dto.AlertDTO, len(alerts))
	for i, alert := range alerts {
		alertDTOs[i] = toAlertDTO(alert)
	}

	return alertDTOs, total, nil
}

// AcknowledgeAlert acknowledges an alert
func (s *AlertService) AcknowledgeAlert(ctx context.Context, tenantID, alertID, userID uuid.UUID) error {
	if err := s.alertRepo.Acknowledge(ctx, tenantID, alertID, userID); err != nil {
		s.logger.Error("Failed to acknowledge alert", zap.Error(err))
		return fmt.Errorf("failed to acknowledge alert")
	}

	s.logger.Info("Alert acknowledged successfully", zap.String("alert_id", alertID.String()))

	return nil
}

// toAlertDTO converts an Alert model to AlertDTO
func toAlertDTO(alert *models.Alert) dto.AlertDTO {
	return dto.AlertDTO{
		ID:             alert.ID,
		TenantID:       alert.TenantID,
		Name:           alert.Name,
		Description:    alert.Description,
		Severity:       alert.Severity,
		Status:         alert.Status,
		TargetType:     alert.TargetType,
		TargetID:       alert.TargetID,
		TriggeredAt:    alert.TriggeredAt,
		AcknowledgedAt: alert.AcknowledgedAt,
		AcknowledgedBy: alert.AcknowledgedBy,
		ResolvedAt:     alert.ResolvedAt,
	}
}
