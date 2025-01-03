package repository

import (
	"infra/api/internal/domain"

	"gorm.io/gorm"
)

type MerchantsRepo struct {
}

func InitMerchantsRepo() *MerchantsRepo {
	return &MerchantsRepo{}
}

func (r *MerchantsRepo) FindByID(tx *gorm.DB, merchantID string) (*domain.Merchants, error) {
	var merchant domain.Merchants
	return &merchant, tx.Where(&domain.Merchants{MerchantID: merchantID}).First(&merchant).Error
}

func (r *MerchantsRepo) FindByName(tx *gorm.DB, merchantName string) (*domain.Merchants, error) {
	var merchant domain.Merchants
	return &merchant, tx.Where(&domain.Merchants{MerchantName: merchantName}).First(&merchant).Error
}

func (r *MerchantsRepo) FindByApiKey(tx *gorm.DB, apiKey string) (*domain.Merchants, error) {
	var merchant domain.Merchants

	return &merchant, tx.Where(&domain.Merchants{ApiKey: apiKey}).First(&merchant).Error
}

func (r *MerchantsRepo) Create(tx *gorm.DB, merchant *domain.Merchants) error {
	return tx.Create(merchant).Error
}
