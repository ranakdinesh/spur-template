package ports

import (
	"context"
	"time"
)

type FunnelStat struct {
	Stage string
	Count int64
}

type SourceStat struct {
	Source     string
	TotalLeads int64
	AvgScore   float64
}

type ReportingRepository interface {
	GetFunnelStats(ctx context.Context, start, end time.Time) ([]FunnelStat, error)
	GetSourceStats(ctx context.Context) ([]SourceStat, error)
}
