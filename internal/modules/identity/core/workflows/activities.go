package workflows

import (
	"context"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
)

type VerificationActivities struct {
	UserRepo ports.UserRepo
	CommPort ports.CommunicationPort
}

func (a *VerificationActivities) CheckVerificationStatus(ctx context.Context, input VerificationReminderInput) (struct{ IsVerified bool }, error) {
	userID, _ := uuid.Parse(input.UserID)
	tenantID, _ := uuid.Parse(input.TenantID)

	user, err := a.UserRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return struct{ IsVerified bool }{false}, err
	}

	return struct{ IsVerified bool }{user.IsVerified()}, nil
}

func (a *VerificationActivities) SendVerificationReminder(ctx context.Context, input VerificationReminderInput) error {
	userID, _ := uuid.Parse(input.UserID)
	tenantID, _ := uuid.Parse(input.TenantID)

	user, err := a.UserRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	// For now, using CommPort to send a placeholder reminder
	// In a real system, we'd have a specific SendVerificationReminder method in CommPort
	// For this task, let's assume SendOTP or a generic Notify works.
	// Since CommPort only has SendOTP for now:
	// return a.CommPort.SendOTP(ctx, user.Email, "email", "REMINDER_PLACEHOLDER")

	// Let's assume we use email if available
	return a.CommPort.SendOTP(ctx, user.Email, "email", "VERIFICATION_REMINDER")
}
