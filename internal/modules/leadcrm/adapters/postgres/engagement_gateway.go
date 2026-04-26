package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
)

type EngagementGatewayStub struct{}

func NewEngagementGateway() ports.EngagementGateway {
	return &EngagementGatewayStub{}
}

func (g *EngagementGatewayStub) EnqueueEngagement(ctx context.Context, lead *domain.Lead, instruction string) (string, error) {
	// Stub implementation - in real world this calls Temporal client
	return "job-123", nil
}
