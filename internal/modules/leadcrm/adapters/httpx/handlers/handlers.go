package handlers

import (
	"github.com/spurbase/spur/internal/modules/leadcrm/core/services"
)

type Handlers struct {
	Services *services.Services
	Reports  *ReportsHandler
}

func New(s *services.Services, reports *ReportsHandler) *Handlers {
	return &Handlers{
		Services: s,
		Reports:  reports,
	}
}
