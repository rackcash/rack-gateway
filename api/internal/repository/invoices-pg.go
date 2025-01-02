package repository

import (
	"infra/api/internal/domain"

	"gorm.io/gorm"
)

type InvoicesRepo struct {
}

func InitInvoicesRepo() *InvoicesRepo {
	return &InvoicesRepo{}
}

func (r *InvoicesRepo) Create(tx *gorm.DB, invoice *domain.Invoices) error {
	return tx.Create(invoice).Error
}

func (r *InvoicesRepo) Update(tx *gorm.DB, invoice *domain.Invoices) error {
	return tx.Save(invoice).Error
}

func (r *InvoicesRepo) FindByID(tx *gorm.DB, invoiceId string) (*domain.Invoices, error) {
	var invoices domain.Invoices
	return &invoices, tx.Where(&domain.Invoices{InvoiceID: invoiceId}).First(&invoices).Error
}
