package service

import (
	"context"
	"infra/api/internal/config"
	"infra/api/internal/domain"
	"infra/api/internal/infra/cache"
	"infra/api/internal/infra/nats"
	"infra/api/internal/logger"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Merchants interface {
	FindByID(tx *gorm.DB, merchantID string) (*domain.Merchants, error)
	FindByApiKey(tx *gorm.DB, apiKey string) (*domain.Merchants, error)
	FindByName(tx *gorm.DB, merchantName string) (*domain.Merchants, error)
	Create(tx *gorm.DB, merchant *domain.Merchants) error
}

type Invoices interface {
	Create(tx *gorm.DB, invoice *domain.Invoices) error
	Update(tx *gorm.DB, invoice *domain.Invoices) error
	FindByID(tx *gorm.DB, invoiceId string) (*domain.Invoices, error)
	FindAndSaveToCache(invoiceId string) (*domain.Invoices, error)
	CalculateFinAmount(amount, rate decimal.Decimal, ceil ...int32) decimal.Decimal
	RunCheck(ctx context.Context, cancel context.CancelFunc, invoice *domain.Invoices, tempWalletAddr string)
	// updates invoice and saves it to cache and db
	UpdateAndSave(tx *gorm.DB, invoice *domain.Invoices) error
	// Tries to find from cache, if not found, searches the database
	FindGlobal(tx *gorm.DB, invoiceId string) (*domain.Invoices, error)

	// for autostart only
	RunFindEnd()
	RunAutostartCheck()
}

type Wallets interface {
	FindByInvoiceID(tx *gorm.DB, invoiceID string) (*domain.Wallets, error)
	FindByMerchantID(tx *gorm.DB, merchantId string, crypto string) (*domain.Wallets, error)
	Create(tx *gorm.DB, wallet *domain.Wallets) error
	CreateAndSave(invoiceId string, merchantId string, crypto domain.Crypto) (*domain.Wallets, error)
}
type Balances interface {
	Create(tx *gorm.DB, balance *domain.Balances) error
	Find(tx *gorm.DB, merchantID, currency string) (*domain.Balances, error)
	FindByPrivate(tx *gorm.DB, private string) (*domain.Balances, error)
	// sends nats message to create wallets (temp wallets) and saves to db
	Init(merchant *domain.Merchants) error
}

type QrCodes interface {
	// generates qr code and saves it to cache
	New(content string) (string, error)
	// returns qr code from cache or generates new one
	FindOrNew(content string) (string, error)
}

type Rates interface {
	Get(amountCurrency string) (*natsdomain.Rates, error)
}
type Locker interface {
	Lock(key string)
	Unlock(key string)
	IsLocked(key string) bool
}

type GetWithdrawal interface {
	StartWaitStatus()
	SetPaidStatus(tx *gorm.DB, balance *domain.Balances, invoice *domain.Invoices, paymentStatus domain.Status) error
}

type GetMerchantWithdrawal interface {
	StartWaitStatus()
}

type OutboxEvents interface {
	StartProcessEvents()
}
type WebhookSender interface {
	Send(url string, info domain.ResponseInvoiceInfo) error
	UpdateList(proxies []string)
	GetList() []string
}

type Withdrawals interface {
	Create(tx *gorm.DB, withdrawal *domain.Withdrawals) error
	Find(tx *gorm.DB, withdrawalId string) (*domain.Withdrawals, error)
}

type Services struct {
	// TODO: Autostart
	OutboxEvents          OutboxEvents
	GetWithdrawal         GetWithdrawal
	GetMerchantWithdrawal GetMerchantWithdrawal
	// CpConfigs             CpConfigs
	Merchants     Merchants
	Invoices      Invoices
	Wallets       Wallets
	Balances      Balances
	QrCodes       QrCodes
	Rates         Rates
	WebhookSender WebhookSender
	Withdrawals   Withdrawals
}

// TODO: code
func HewServices(ns *natsdomain.Ns, db *gorm.DB, l logger.Logger, config *config.Config) *Services {
	n := &nats.NatsInfra{Ns: ns}

	walletsRepo := repository.InitWalletsRepo()
	balancesRepo := repository.InitBalancesRepo()
	lockerService := NewLockerService(cache.InitStorage())

	webhookSender := NewWebhookSenderService(config.ProxyList, l)

	invoiceService := NewInvoicesService(db, repository.InitInvoicesRepo(), walletsRepo, balancesRepo, lockerService, n, l, cache.InitStorage(), config)

	eventsRepo := repository.InitEventsRepo()
	GetWithdrawalService := NewGetWithdrawalService(db, n, l, eventsRepo, walletsRepo, balancesRepo, invoiceService, webhookSender, config)

	withdrawalsRepo := repository.InitWithdrawalsRepo()

	return &Services{
		GetMerchantWithdrawal: NewGetMerchantWithdrawalService(db, n, l, balancesRepo, withdrawalsRepo, config),
		WebhookSender:         webhookSender,
		OutboxEvents:          NewOutboxEventsService(invoiceService, balancesRepo, walletsRepo, GetWithdrawalService, n, db, l, eventsRepo, webhookSender),
		GetWithdrawal:         GetWithdrawalService,
		// CpConfigs:             NewCpConfigsService(db, repository.InitCpConfigsRepo()),
		Merchants:   NewMerchantsService(db, repository.InitMerchantsRepo(), ns),
		Invoices:    invoiceService,
		Wallets:     NewWalletsService(db, walletsRepo, ns),
		Balances:    NewBalancesService(db, balancesRepo, ns),
		QrCodes:     NewQrCodesService(),
		Rates:       NewRatesService(cache.InitStorage(), ns),
		Withdrawals: NewWithdrawalService(db, withdrawalsRepo, ns),
	}
}
