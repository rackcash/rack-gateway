package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type responseError struct {
	Error   bool   `json:"error"`
	ErrorID string `json:"error_id"`
	Msg     string `json:"msg"`
}

type responseInvoiceCreatedWallet struct {
	QrCode         string          `json:"qr_code"`
	Address        string          `json:"address"`
	AmountToPay    decimal.Decimal `json:"amount_to_pay"`
	Cryptocurrency string          `json:"cryptocurrency"`
}

type responseWithdrawalStarted struct {
	Error        bool   `json:"error"`
	WithdrawalID string `json:"withdrawal_id"`
	ToAddress    string `json:"to_address"`
	Amount       string `json:"amount"`
	Status       string `json:"info"`
}

type responseWithdrawalInfo struct {
	Error     bool   `json:"error"`
	ToAddress string `json:"to_address"`
	Amount    string `json:"amount"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// /currency/convert
type responseConverterOK struct {
	Error          bool            `json:"error"`
	Fiat           string          `json:"fiat"`
	Amount         decimal.Decimal `json:"amount"`
	Cryptocurrency string          `json:"cryptocurrency"`
	Converted      decimal.Decimal `json:"converted"`
	Rate           decimal.Decimal `json:"rate"`
}

type responseRates struct {
	Eth decimal.Decimal `json:"eth"`
	Ltc decimal.Decimal `json:"ltc"`
	Sol decimal.Decimal `json:"sol"`
	Ton decimal.Decimal `json:"ton"`
}

// /currency/rates
type responseRatesOK struct {
	Error bool          `json:"error"`
	Fiat  string        `json:"fiat"`
	Rates responseRates `json:"rates"`
}

// /invoice/create
type responseInvoiceCreatedInfo struct {
	Id string `json:"id"`
	// Amount   decimal.Decimal `json:"amount"`
	// Currency string                       `json:"currency"`
	Wallet responseInvoiceCreatedWallet `json:"wallet"`
}

type responseInvoiceCreated struct {
	Error   bool                       `json:"error"`
	Invoice responseInvoiceCreatedInfo `json:"invoice"`
}

type responseMerchantCreated struct {
	Error      bool   `json:"error"`
	ApiKey     string `json:"api_key"`
	MerchantId string `json:"merchant_id"`
}

type responseInvoiceCancelled struct {
	Error bool `json:"error"`
	// Message string `json:"message"`
}

func responseErr(c *gin.Context, statusCode int, msg, errorID string) {
	// var errorId string

	// if _errorId == "" {
	// 	errorId = "N/A"
	// }

	c.AbortWithStatusJSON(statusCode, responseError{true, errorID, msg})
}
