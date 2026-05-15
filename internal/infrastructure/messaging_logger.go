package infrastructure

import (
	"context"

	"github.com/ranakdinesh/spur-template/internal/infrastructure/messaging"
	"github.com/ranakdinesh/spur-template/internal/logger"
)

type messagingLogger struct {
	log *logger.Loggerx
}

func (l messagingLogger) WarnMessagingNotConfigured(ctx context.Context, operation string, channel messaging.Channel, err error) {
	event := l.log.Warn(ctx).
		Str("component", "messaging_gateway").
		Str("operation", operation).
		Err(err)
	if channel != "" {
		event = event.Str("channel", string(channel))
	}
	event.Msg("messaging service is not configured; skipping message dispatch")
}
