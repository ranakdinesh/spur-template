package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type VerificationReminderInput struct {
	UserID   string
	TenantID string
}

func VerificationReminderWorkflow(ctx workflow.Context, input VerificationReminderInput) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// Reminders at 3, 7, 10, 14 days
	// Using relative delays
	delays := []time.Duration{
		3 * 24 * time.Hour,
		4 * 24 * time.Hour, // 7 days total
		3 * 24 * time.Hour, // 10 days total
		4 * 24 * time.Hour, // 14 days total
	}

	for _, delay := range delays {
		err := workflow.Sleep(ctx, delay)
		if err != nil {
			return err
		}

		var status struct {
			IsVerified bool
		}
		err = workflow.ExecuteActivity(ctx, "CheckVerificationStatus", input).Get(ctx, &status)
		if err != nil {
			return err
		}

		if status.IsVerified {
			return nil // User verified, stop workflow
		}

		// Send reminder
		err = workflow.ExecuteActivity(ctx, "SendVerificationReminder", input).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
