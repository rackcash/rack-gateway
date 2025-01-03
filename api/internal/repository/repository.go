package repository

import (
	"infra/api/internal/domain"

	"gorm.io/gorm"
)

type Merchants interface {
	FindByID(tx *gorm.DB, merchantID string) (*domain.Merchants, error)
	FindByApiKey(tx *gorm.DB, apiKey string) (*domain.Merchants, error)
	FindByName(tx *gorm.DB, merchantName string) (*domain.Merchants, error)
	Create(tx *gorm.DB, merchant *domain.Merchants) error
}

type Invoices interface {
	Create(tx *gorm.DB, invoice *domain.Invoices) error
	Update(tx *gorm.DB, invoice *domain.Invoices) error
	FindByID(tx *gorm.DB, invoiceId string) (*domain.Invoices, error)
}

type Wallets interface {
	FindByInvoiceID(tx *gorm.DB, invoiceID string) (*domain.Wallets, error)
	FindByMerchantID(tx *gorm.DB, merchantID string, crypto string) (*domain.Wallets, error)
	Create(tx *gorm.DB, wallet *domain.Wallets) error
}

type Balances interface {
	Create(tx *gorm.DB, balance *domain.Balances) error
	Find(tx *gorm.DB, merchantID, currency string) (*domain.Balances, error)
	FindByPrivate(tx *gorm.DB, private string) (*domain.Balances, error)
}

type Events interface {
	Create(tx *gorm.DB, eventType string, eventRelationID uint, payload string) error
	Done(tx *gorm.DB, eventRelationID uint, eventType string) error
	Find(tx *gorm.DB, eventRelationID uint, eventType string) (*domain.Events, error)
}

type Withdrawals interface {
	Create(tx *gorm.DB, withdrawal *domain.Withdrawals) error
	Find(tx *gorm.DB, withdrawalId string) (*domain.Withdrawals, error)
	UpdateStatus(tx *gorm.DB, withdrawalId string, status domain.WithdrawalStatus) error
}

type Repositories struct {
	Merchants   Merchants
	Invoices    Invoices
	Wallets     Wallets
	Balances    Balances
	Events      Events
	Withdrawals Withdrawals
}

func New() *Repositories {
	return &Repositories{
		Events:      InitEventsRepo(),
		Merchants:   InitMerchantsRepo(),
		Invoices:    InitInvoicesRepo(),
		Wallets:     InitWalletsRepo(),
		Balances:    InitBalancesRepo(),
		Withdrawals: InitWithdrawalsRepo(),
	}
}
