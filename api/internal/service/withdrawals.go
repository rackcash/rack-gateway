package service

import (
	"infra/api/internal/domain"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"

	"gorm.io/gorm"
)

type WithdrawalService struct {
	repo repository.Withdrawals
	db   *gorm.DB
	ns   *natsdomain.Ns
}

func NewWithdrawalService(db *gorm.DB, repo repository.Withdrawals, ns *natsdomain.Ns) *WithdrawalService {
	return &WithdrawalService{db: db, repo: repo, ns: ns}
}

func (s *WithdrawalService) Create(tx *gorm.DB, withdrawal *domain.Withdrawals) error {
	return s.repo.Create(tx, withdrawal)
}

func (s *WithdrawalService) Find(tx *gorm.DB, withdrawalId string) (*domain.Withdrawals, error) {
	return s.repo.Find(tx, withdrawalId)
}
