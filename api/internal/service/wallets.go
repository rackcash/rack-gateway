package service

import (
	"encoding/json"
	"errors"
	"infra/api/internal/domain"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletsService struct {
	repo repository.Wallets
	db   *gorm.DB
	ns   *natsdomain.Ns
}

func NewWalletsService(db *gorm.DB, repo repository.Wallets, ns *natsdomain.Ns) *WalletsService {
	return &WalletsService{db: db, repo: repo, ns: ns}
}

func (s *WalletsService) FindByInvoiceID(tx *gorm.DB, invoiceID string) (*domain.Wallets, error) {
	return s.repo.FindByInvoiceID(tx, invoiceID)
}

func (s *WalletsService) Create(tx *gorm.DB, wallet *domain.Wallets) error {
	return s.repo.Create(tx, wallet)
}

// wrapper for create and nats req
func (s *WalletsService) CreateAndSave(invoiceId string, merchantId string, crypto domain.Crypto) (*domain.Wallets, error) {
	var wallet = &domain.Wallets{
		InvoiceID: invoiceId,
		// Address: ,
		// Private: ,
		MerchantID: merchantId,

		Balance: decimal.NewFromInt(0),
		Crypto:  crypto.ToString(),
	}

	jsonMsg, err := json.Marshal(natsdomain.ReqNewWallet{Cryptocurrency: crypto.ToString()})
	if err != nil {
		return nil, err
	}

	dataBytes, err := s.ns.ReqAndRecv(natsdomain.SubjNewWallet, jsonMsg)
	if err != nil {
		return nil, err
	}

	newWallet, err := utils.Unmarshal[natsdomain.ResNewWallet](dataBytes)
	if err != nil {
		return nil, err
	}

	wallet.Address = newWallet.Address
	wallet.Private = newWallet.PrivateKey

	w, err := s.repo.FindByInvoiceID(s.db, wallet.InvoiceID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// not found

		err := s.repo.Create(s.db, wallet)
		if err != nil {
			return nil, err
		} else {
			return wallet, nil
		}
	}

	return w, nil
}
