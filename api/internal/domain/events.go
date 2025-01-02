package domain

import "time"

const (
	EVENT_INVOICE_PROCESSING = "invoice_processing" // sending from temp wallet to the main wallet
	EVENT_WEBHOOK            = "webhook"
)

type Events struct {
	ID         uint   `gorm:"primaryKey"`
	RelationID uint   `gorm:"not null"`
	Type       string `gorm:"type:varchar(255)"` //const type Event*
	Payload    string
	Status     string // new/done
	CreatedAt  time.Time
}

// event payloads
type PayloadInvoiceProcessing struct {
	InvoiceID      string `json:"invoice_id"`
	TxHash         string `json:"tx_hash"`
	TxTempWallet   string `json:"tx_temp_wallet"`  // temp wallet address
	BalanceAddress string `json:"balance_address"` // this is the address where the money is sent after processing (from tx_temp_wallet)
	CryptoAmount   string `json:"crypto_amount"`
	Status         string `json:"status"`
}

type WebhookPayload struct {
	MerchantID string              `json:"merchant_id"`
	InvoiceID  string              `json:"invoice_id"`
	Url        string              `json:"url"`
	Info       ResponseInvoiceInfo `json:"info"`
}
