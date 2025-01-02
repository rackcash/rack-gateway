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
	Error      bool
	ApiKey     string `json:"api_key"`
	MerchantId string `json:"merchant_id"`
}

func responseErr(c *gin.Context, statusCode int, msg, errorID string) {
	// var errorId string

	// if _errorId == "" {
	// 	errorId = "N/A"
	// }

	c.AbortWithStatusJSON(statusCode, responseError{true, errorID, msg})
}
