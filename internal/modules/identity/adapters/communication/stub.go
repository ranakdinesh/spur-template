package communication

import (
	"context"

	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/platform/logger"
)

type StubCommunicationService struct {
	log *logger.Loggerx
}

func NewStubCommunicationService(log *logger.Loggerx) ports.CommunicationPort {
	return &StubCommunicationService{
		log: log,
	}
}

func (s *StubCommunicationService) SendOTP(ctx context.Context, recipient string, channel string, code string) error {
	s.log.Info(ctx).
		Str("channel", channel).
		Str("recipient", recipient).
		Str("otp_code", code).
		Msg("STUB: Sending OTP")
	return nil
}
