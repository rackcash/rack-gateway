package service

import (
	"infra/api/internal/domain"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"

	"gorm.io/gorm"
)

type MerchantsService struct {
	repo repository.Merchants
	db   *gorm.DB
	ns   *natsdomain.Ns
}

func NewMerchantsService(db *gorm.DB, repo repository.Merchants, ns *natsdomain.Ns) *MerchantsService {
	return &MerchantsService{repo: repo, ns: ns, db: db}
}

func (s *MerchantsService) FindByID(tx *gorm.DB, merchantID string) (*domain.Merchants, error) {
	return s.repo.FindByID(tx, merchantID)
}

func (s *MerchantsService) FindByApiKey(tx *gorm.DB, apiKey string) (*domain.Merchants, error) {
	return s.repo.FindByApiKey(tx, apiKey)
}

func (s *MerchantsService) FindByName(tx *gorm.DB, apiKey string) (*domain.Merchants, error) {
	return s.repo.FindByName(tx, apiKey)
}

func (s *MerchantsService) Create(tx *gorm.DB, merchant *domain.Merchants) error {
	return s.repo.Create(tx, merchant)
}

func (s *MerchantsService) ApiKeyExists(tx *gorm.DB, apiKey string) (bool, error) {
	_, err := s.FindByApiKey(tx, apiKey)
	if err != nil {
		if postgres.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
