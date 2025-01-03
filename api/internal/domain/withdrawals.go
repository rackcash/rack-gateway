package domain

import "github.com/shopspring/decimal"

type WithdrawalStatus uint8

const (
	WITHDRAWAL_NONE WithdrawalStatus = iota // only for init
	WITHDRAWAL_PROCESSING
	WITHDRAWAL_SUCCESS
	WITHDRAWAL_ERROR
)

var WithdrawalStatuses = [...]string{"none", "processing", "success", "error"}

func (ws WithdrawalStatus) ToString() string {
	return WithdrawalStatuses[ws]
}
func (ws WithdrawalStatus) End() bool {
	return ws == WITHDRAWAL_ERROR || ws == WITHDRAWAL_SUCCESS
}

type Withdrawals struct {
	Model
	ID uint `gorm:"primaryKey"`

	WithdrawalID string           `gorm:"unique;not null"`
	Amount       decimal.Decimal  `gorm:"type:numeric"`
	From         string           `gorm:"not null"`
	To           string           `gorm:"not null"`
	Crypto       string           `gorm:"not null"`
	Status       WithdrawalStatus `gorm:"not null"`
}
