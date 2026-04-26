package services

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/domain"
	"github.com/spurbase/spur/internal/modules/leadcrm/core/ports"
	"context"
	"time"
)

type ConversionService struct {
	leadRepo     ports.LeadRepository
	contactRepo  ports.ContactRepository
	accountRepo  ports.AccountRepository
	activityRepo ports.ActivityRepository
	tm           ports.TransactionManager
}

func NewConversionService(l ports.LeadRepository, c ports.ContactRepository, a ports.AccountRepository, act ports.ActivityRepository, tm ports.TransactionManager) *ConversionService {
	return &ConversionService{
		leadRepo:     l,
		contactRepo:  c,
		accountRepo:  a,
		activityRepo: act,
		tm:           tm,
	}
}

func (s *ConversionService) Convert(ctx context.Context, leadID string, accountName string) error {
	return s.tm.RunInTx(ctx, func(ctx context.Context) error {
		// 1. Lock Lead
		lead, err := s.leadRepo.GetForUpdate(ctx, leadID)
		if err != nil {
			return err // Handle not found
		}

		if lead.Status == domain.LeadStatusConverted {
			return nil // Already converted
		}

		// 2. Create Account
		if accountName == "" {
			accountName = "Unknown Account" // or derive from lead company
		}

		// Use repo to save (stub IDs)
		acct := &domain.Account{
			ID:        "gen-acct-" + leadID,
			TenantID:  lead.TenantID,
			Name:      accountName,
			CreatedAt: time.Now(),
		}
		if err := s.accountRepo.Save(ctx, acct); err != nil {
			return err
		}

		// 3. Create Contact
		contact := &domain.Contact{
			ID:       "gen-cont-" + leadID,
			TenantID: lead.TenantID,
			LeadID:   &lead.ID,
			// Naive name split stub
			FirstName: "Lead",
			LastName:  "Contact",
			CreatedAt: time.Now(),
		}
		if err := s.contactRepo.Save(ctx, contact); err != nil {
			return err
		}

		// 4. Update Lead
		lead.Status = domain.LeadStatusConverted
		lead.Stage = domain.StageConvert
		lead.StageChangedAt = time.Now()
		if err := s.leadRepo.Save(ctx, lead); err != nil {
			return err
		}

		// 5. Emit Activity
		_ = s.activityRepo.Save(ctx, &domain.Activity{
			ID:        "act-conv-" + lead.ID,
			TenantID:  lead.TenantID,
			Type:      domain.ActivityTypeConversion,
			Details:   "Converted to Account: " + acct.Name,
			CreatedAt: time.Now(),
		})

		return nil
	})
}
