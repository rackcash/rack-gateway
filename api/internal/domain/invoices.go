package domain

import (
	"github.com/shopspring/decimal"
)

type Invoices struct {
	Model
	ID           uint   `gorm:"primaryKey"`
	InvoiceID    string `gorm:"unique;not null"`
	MerchantID   string `gorm:"size:36;not null"`
	Status       Status `gorm:"type:int8"`
	EndTimestamp int64  // unix invoice end timestamp
	// CryptoAmount   decimal.Decimal `gorm:"type:numeric"` // final amount to be paid in the desired cryptocurrency
	Cryptocurrency string          `gorm:"type:text"` // after selecting the desired cryptocurrency (frontend)
	Amount         decimal.Decimal `gorm:"type:numeric"`
	// AmountCurrency  string
	ProcessedTxHash string `gorm:"type:text"`          // After the client sends money to the temp wallet, the server sends this money to the main wallet. this is the hash of the transaction of sending to the main wallet
	Webhook         string `gorm:"type:text;not null"` // webhook url. used in webhook sender service
}

type Status uint8

const (
	STATUS_NOT_PAID Status = iota
	STATUS_PAID
	STATUS_PAID_OVER
	STATUS_END
	STATUS_PAID_LESS
	STATUS_IN_PROCESSING
	STATUS_CANCELLED
)

var Statuses = [...]string{"not_paid", "paid", "paid_over", "end", "paid_less", "processing", "cancelled"}

// methods

func StrToStatus(s string) Status {
	for i, statusName := range Statuses {
		if s == statusName {
			return Status(i)
		}
	}
	return STATUS_NOT_PAID

}

func (s Status) ToString() string {
	return Statuses[s]
}

func (s Status) IsCancelled() bool {
	return s == STATUS_CANCELLED
}
func (s Status) IsPaid() bool {
	return s == STATUS_PAID || s == STATUS_PAID_OVER
}

func (i *Invoices) IsInProcessing() bool {
	return i.Status == STATUS_IN_PROCESSING
}

func (s Status) IsNotPaid() bool {
	return s == STATUS_NOT_PAID
}
