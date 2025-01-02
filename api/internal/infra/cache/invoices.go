package cache

import (
	"infra/api/internal/domain"
)

func SaveInvoice(invoiceId string, invoice *domain.Invoices) {
	InvoiceMap.Store(invoiceId, invoice)
}

func FindInvoice(invoiceId string) *domain.Invoices {
	v, ok := InvoiceMap.Load(invoiceId)
	if !ok {
		return nil
	}

	return v.(*domain.Invoices)
}

// updates the invoice validation status to avoid parallel validation of the same invoice
//
// used in helpers/invoice.go -> CheckInvoice()
func SetInvoiceBusy(invoiceId string, busy bool) {
	InvoiceCheckMap.Store(invoiceId, busy)
}

func SetInvoicePaymentStatus(invoiceId string, invoice *domain.Invoices, status domain.Status) {
	invoice.Status = status
	InvoiceMap.Store(invoiceId, invoice)
}

func IsInvoiceBusy(invoiceId string) bool {
	v, ok := InvoiceCheckMap.Load(invoiceId)
	if !ok {
		return false
	}

	return v.(bool)
}
