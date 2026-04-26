package ports

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"context"
)

type SourceIngestor interface {
	Ingest(ctx context.Context, source domain.Source, payload []byte) error
}
