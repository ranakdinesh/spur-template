package ports

import (
	"context"
)

type CommunicationPort interface {
	SendOTP(ctx context.Context, recipient string, channel string, code string) error
}
