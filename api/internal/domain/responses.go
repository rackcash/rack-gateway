package domain

import (
	"errors"
	"fmt"
	"net/http"
)

type ResponseInvoiceInfo struct {
	Id     string `json:"id"`
	Amount string `json:"amount"`
	// Currency       string `json:"currency"`
	// CryptoAmount   string `json:"crypto_amount"`
	Cryptocurrency string `json:"cryptocurrency"`
	Status         string `json:"status"`
	IsPaid         bool   `json:"is_paid"`
	CreatedAt      string `json:"created_at"`
}

const (
	ErrMsgRateLimitExceeded         = "rate limit exceeded"
	ErrMsgInternalServerError       = "internal server error"
	ErrMsgParamsInternalServerError = "internal server error: %s"
	ErrMsgBadRequest                = "bad request"
	ErrMsgParamsBadRequest          = "bad request: %s"
	ErrMsgAccessError               = "access error"

	ErrMsgMerchantIdExists   = "merchant with that id already exists"
	ErrMsgMerchantNameExists = "merchant with that name already exists"

	ErrMsgMerchantNotFound = "merchant not found"

	ErrMsgInsufficientFundsParams = "insufficient funds. available: %s"

	ErrMsgWithdrawalNotFound = "withdrawal not found"

	ErrMsgGetBalanceError = "get balance error"

	ErrMsgSecurityError         = "security error"
	ErrMsgApiKeyNotFound        = "api key not found"
	ErrMsgApiKeyInvalid         = "invalid api key"
	ErrMsgInvalidInvoiceId      = "invalid invoice id"
	ErrConfigNotFound           = "config not found"
	ErrMsgCryptoAlreadySelected = "cryptocurrency already selected"
	ErrMsgInvalidCrypto         = "invalid cryptocurrency"
	ErrMsgInitBalancesError     = "can't init balances"
)

var (
	ErrInvalidInvoiceId    = fmt.Errorf("invalid invoice id")
	Err                    = fmt.Errorf("invalid invoice id")
	ErrInternalServerError = fmt.Errorf(ErrMsgInternalServerError)
	ErrInvoiceIdNotFound   = fmt.Errorf("invoice id not found")
)

const (
	ErrParamEmptyInvoiceId = "invoice id is empty"
)

func GetStatusByErr(err error) (status int) {
	if err == nil {
		return 200
	}

	switch {
	case errors.Is(err, ErrInternalServerError):
		status = http.StatusInternalServerError
	case errors.Is(err, ErrInvalidInvoiceId):
		status = http.StatusBadRequest
	case errors.Is(err, ErrInvoiceIdNotFound):
		status = http.StatusBadRequest
	default:
		status = http.StatusInternalServerError
	}
	return status
}
