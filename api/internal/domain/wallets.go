package domain

import "github.com/shopspring/decimal"

type Wallets struct {
	Model
	InvoiceID  string          `gorm:"not null"`
	MerchantID string          `gorm:"size:36;not null"`
	Address    string          `gorm:"not null"`
	Private    string          // private key or seed phrase
	Balance    decimal.Decimal `gorm:"type:numeric;default:0"`
	Crypto     string          `gorm:"type:text"`
}
