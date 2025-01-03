package repository

import (
	"infra/api/internal/domain"

	"gorm.io/gorm"
)

type BalancesRepo struct {
}

func InitBalancesRepo() *BalancesRepo {
	return &BalancesRepo{}
}

func (r *BalancesRepo) Create(tx *gorm.DB, balance *domain.Balances) error {
	return tx.Create(balance).Error
}

func (r *BalancesRepo) Find(tx *gorm.DB, merchantID, currency string) (*domain.Balances, error) {
	var balance domain.Balances
	return &balance, tx.Where(&domain.Balances{MerchantID: merchantID, Crypto: currency}).First(&balance).Error
}

func (r *BalancesRepo) FindByPrivate(tx *gorm.DB, private string) (*domain.Balances, error) {
	var balance domain.Balances
	return &balance, tx.Where(&domain.Balances{Private: private}).First(&balance).Error
}
