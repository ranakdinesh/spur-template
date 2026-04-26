package services

import "github.com/spurbase/spur/internal/modules/leadcrm/core/ports"

type Services struct {
	Lead          ports.LeadService
	Qualification ports.QualificationService
	Conversion    ports.ConversionService
}

func New(l ports.LeadService, q ports.QualificationService, c ports.ConversionService) *Services {
	return &Services{
		Lead:          l,
		Qualification: q,
		Conversion:    c,
	}
}
