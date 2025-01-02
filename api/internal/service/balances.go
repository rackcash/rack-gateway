package service

import (
	"infra/api/internal/domain"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BalancesService struct {
	repo repository.Balances
	db   *gorm.DB
	ns   *natsdomain.Ns
}

func NewBalancesService(db *gorm.DB, repo repository.Balances, ns *natsdomain.Ns) *BalancesService {
	return &BalancesService{repo: repo, ns: ns, db: db}
}

func (s *BalancesService) Create(tx *gorm.DB, balance *domain.Balances) error {
	return s.repo.Create(tx, balance)
}

func (s *BalancesService) Find(tx *gorm.DB, merchantID, currency string) (*domain.Balances, error) {
	return s.repo.Find(tx, merchantID, currency)
}

func (s *BalancesService) FindByPrivate(tx *gorm.DB, private string) (*domain.Balances, error) {
	return s.repo.FindByPrivate(tx, private)
}

func (s *BalancesService) Init(merchant *domain.Merchants) error {
	// TODO: add more cryptocurrencies
	balances := []domain.Balances{
		{MerchantID: merchant.MerchantID, Balance: decimal.Zero, Crypto: domain.CRYPTO_ETH.ToString()},
		{MerchantID: merchant.MerchantID, Balance: decimal.Zero, Crypto: domain.CRYPTO_TON.ToString()},
		{MerchantID: merchant.MerchantID, Balance: decimal.Zero, Crypto: domain.CRYPTO_SOL.ToString()},
	}

	for _, balance := range balances {

		_, err := s.repo.Find(s.db, balance.MerchantID, balance.Crypto)
		if err == nil {
			continue
		}

		if !postgres.IsNotFound(err) {
			return err
		}

		address, private, err := createWallet(s.ns, balance.Crypto)
		if err != nil {
			return err
		}

		balance.Address = address
		balance.Private = private

		if s.repo.Create(s.db, &balance) != nil {
			return err
		}
	}

	return nil

}

func createWallet(ns *natsdomain.Ns, currency string) (address, private string, err error) {
	data, err := ns.ReqAndRecv(natsdomain.SubjNewWallet, utils.MustMarshal(natsdomain.ReqNewWallet{Cryptocurrency: currency}))
	if err != nil {
		return "", "", err
	}

	nw, err := utils.Unmarshal[natsdomain.ResNewWallet](data)
	if err != nil {
		return "", "", err
	}

	return nw.Address, nw.PrivateKey, nil
}
