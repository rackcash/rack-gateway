package repository

import (
	"infra/api/internal/domain"

	"gorm.io/gorm"
)

type WalletsRepo struct {
}

func InitWalletsRepo() *WalletsRepo {
	return &WalletsRepo{}
}

func (r *WalletsRepo) FindByInvoiceID(tx *gorm.DB, invoiceID string) (*domain.Wallets, error) {
	var wallet domain.Wallets
	return &wallet, tx.Where(&domain.Wallets{InvoiceID: invoiceID}).First(&wallet).Error
}

func (r *WalletsRepo) Create(tx *gorm.DB, wallet *domain.Wallets) error {
	return tx.Create(wallet).Error
}

func (r *WalletsRepo) FindByMerchantID(tx *gorm.DB, merchantID string, crypto string) (*domain.Wallets, error) {
	var wallet domain.Wallets
	return &wallet, tx.Where(&domain.Wallets{MerchantID: merchantID, Crypto: crypto}).First(&wallet).Error
}
