package postgres

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
	"time"
)

type ReportingRepo struct {
	adapter *Adapter
}

func NewReportingRepo(adapter *Adapter) ports.ReportingRepository {
	return &ReportingRepo{adapter: adapter}
}

func (r *ReportingRepo) GetFunnelStats(ctx context.Context, start, end time.Time) ([]ports.FunnelStat, error) {
	// Stub implementation until SQLC generated code is linked
	// r.adapter.q.GetFunnelStats(...)
	return []ports.FunnelStat{
		{Stage: "capture", Count: 100},
		{Stage: "triage", Count: 50},
		{Stage: "qualify", Count: 20},
		{Stage: "convert", Count: 5},
	}, nil
}

func (r *ReportingRepo) GetSourceStats(ctx context.Context) ([]ports.SourceStat, error) {
	// Stub implementation
	return []ports.SourceStat{
		{Source: "webform", TotalLeads: 120, AvgScore: 45.5},
		{Source: "api", TotalLeads: 30, AvgScore: 20.0},
		{Source: "manual", TotalLeads: 5, AvgScore: 10.0},
	}, nil
}
