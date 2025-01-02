package domain

import "github.com/shopspring/decimal"

type Balances struct {
	Model
	ID         uint            `gorm:"primaryKey"`
	Address    string          `gorm:"not null"`
	Private    string          // private key or seed phrase
	MerchantID string          `gorm:"size:36;not null"`
	Balance    decimal.Decimal `gorm:"type:numeric;default:0"`
	Crypto     string          `gorm:"type:text"`
}
