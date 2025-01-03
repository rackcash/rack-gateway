package repository

import (
	"infra/api/internal/domain"

	"gorm.io/gorm"
)

type WithdrawalsRepo struct {
}

func InitWithdrawalsRepo() *WithdrawalsRepo {
	return &WithdrawalsRepo{}
}

func (r *WithdrawalsRepo) Create(tx *gorm.DB, withdrawal *domain.Withdrawals) error {
	return tx.Create(withdrawal).Error
}

func (r *WithdrawalsRepo) Find(tx *gorm.DB, withdrawalId string) (*domain.Withdrawals, error) {
	var withdrawals domain.Withdrawals
	return &withdrawals, tx.Where(&domain.Withdrawals{WithdrawalID: withdrawalId}).First(&withdrawals).Error
}

func (r *WithdrawalsRepo) UpdateStatus(tx *gorm.DB, withdrawalId string, status domain.WithdrawalStatus) error {
	return tx.Model(&domain.Withdrawals{}).Where(&domain.Withdrawals{WithdrawalID: withdrawalId}).Updates(domain.Withdrawals{Status: status}).Error
}
