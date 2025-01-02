package logger

import (
	"github.com/shopspring/decimal"
)

// func (l Logger) TemplInvoiceLog(logLevel LogLevel, message string, invoiceId string, amount decimal.Decimal, currency string, uri string, ip string) string {
// 	errorId := GenErrorId()

// 	switch logLevel {
// 	case LL_ERROR:
// 		l.Error(message, LS_INVOICES, true, "invoice_id", invoiceId, "amount", amount.String(), "currency", currency, "uri", uri, "error_id", errorId, "ip", ip)
// 	case LL_INFO:
// 		l.Info(message, LS_INVOICES, true, "invoice_id", invoiceId, "amount", amount.String(), "currency", currency, "uri", uri)
// 	case LL_FATAL:
// 		l.Fatal(message, LS_INVOICES, true, "invoice_id", invoiceId, "amount", amount.String(), "currency", currency, "uri", uri)
// 	default:
// 		fmt.Printf("invalid log level: %d\n", logLevel)
// 		return ""
// 	}

// 	return errorId
// }

// TODO: change currency to cryptocurrency
func (l Logger) TemplInvoiceErr(message string, errorId string, invoiceId string, amount decimal.Decimal, currency string, uri string, merchantId string, ip string) string {

	l.Error(message, LS_INVOICES, true, "invoice_id", invoiceId, "amount", amount.String(), "currency", currency, "uri", uri, "error_id", errorId, "ip", ip, "merchant_id", merchantId)
	return errorId
}

func (l Logger) TemplInvoiceInfo(message string, errorId string, invoiceId string, amount decimal.Decimal, currency string, uri string, merchantId string, ip string) string {
	l.Info(message, LS_INVOICES, true, "invoice_id", invoiceId, "amount", amount.String(), "currency", currency, "uri", uri, "error_id", errorId, "ip", ip, "merchant_id", merchantId)
	return errorId
}

// use only for fatal errors
func (l Logger) TemplHTTPError(message string, ipv4 string, err error) {
	l.Fatal(message, LS_FATAL, true, "error", err.Error(), "ipv4", ipv4)
}

func (l Logger) TemplNatsError(message, natsUrl string, err error) {
	l.Error(message, LS_NATS, true, "nats_url", natsUrl, "error", err.Error())
}

func (l Logger) TemplNatsInfo(message, natsUrl string) {
	l.Info(message, LS_NATS, true, "nats_url", natsUrl, "error", "N/A")
}

func (l Logger) TemplWebhookErr(message, url string, attempts int, proxy string, payload []byte) {
	l.Error(message, LS_WEBHOOKS, true, "url", url, "attempts", attempts, "proxy", proxy, "payload", string(payload))
}

func (l Logger) TemplWebhookInfo() {

}
