package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"infra/api/internal/config"
	"infra/api/internal/domain"
	"infra/api/internal/infra/cache"
	"infra/api/internal/infra/nats"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/logger"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type InvoicesService struct {
	repo     repository.Invoices
	wallets  repository.Wallets
	balances repository.Balances
	locker   Locker
	n        *nats.NatsInfra
	db       *gorm.DB
	cache    *cache.Cache
	l        logger.Logger
	config   *config.Config
}

func NewInvoicesService(db *gorm.DB, repo repository.Invoices, wallets repository.Wallets, balances repository.Balances, locker Locker, n *nats.NatsInfra, l logger.Logger, cache *cache.Cache, config *config.Config) *InvoicesService {
	return &InvoicesService{repo: repo, wallets: wallets, balances: balances, n: n, db: db, l: l, cache: cache, locker: locker, config: config}
}

func (s *InvoicesService) Create(tx *gorm.DB, invoice *domain.Invoices) error {
	return s.repo.Create(tx, invoice)
}

func (s *InvoicesService) Update(tx *gorm.DB, invoice *domain.Invoices) error {
	return s.repo.Update(tx, invoice)
}

func (s *InvoicesService) FindByID(tx *gorm.DB, invoiceId string) (*domain.Invoices, error) {
	return s.repo.FindByID(tx, invoiceId)
}

func (s *InvoicesService) FindGlobal(tx *gorm.DB, invoiceId string) (*domain.Invoices, error) {
	// validate uuid (to avoid unnecessary database and cache queries)
	if uuid.Validate(invoiceId) != nil {
		return nil, domain.ErrInvalidInvoiceId
	}

	var errid = logger.GenErrorId()

	//  try to find in cahce
	cacheV := s.cache.Load(invoiceId)
	if cacheV != nil { // found
		return utils.SafeCast[*domain.Invoices](cacheV)
	}

	invoice, err := s.repo.FindByID(s.db, invoiceId)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, domain.ErrInvoiceIdNotFound
		}

		s.l.TemplInvoiceErr("find invoice by id error: "+err.Error(), errid, invoiceId, decimal.Zero, logger.NA, logger.NA, logger.NA, logger.NA)
		return nil, domain.ErrInternalServerError
	}

	if invoice == nil {
		return nil, domain.ErrInternalServerError
	}

	return invoice, nil
}

func (s *InvoicesService) UpdateAndSave(tx *gorm.DB, invoice *domain.Invoices) error {
	// return fmt.Errorf("test")
	err := s.repo.Update(tx, invoice)
	if err != nil {
		return err
	}

	s.cache.Set(invoice.InvoiceID, invoice, time.Minute*5)
	return nil
}

func (s *InvoicesService) FindAndSaveToCache(invoiceId string) (*domain.Invoices, error) {
	invoice, err := s.FindGlobal(s.db, invoiceId)
	if err != nil {
		return nil, err
	}

	s.cache.Set(invoiceId, invoice, time.Minute*5)

	return invoice, nil
}

func (s *InvoicesService) RunCheck(ctx context.Context, cancel context.CancelFunc, invoice *domain.Invoices, tempWalletAddr string) {
	const reconnectDelay = 15 * time.Second // after error wait n secs and try again

	var errid = logger.GenErrorId()

	if s.locker.IsLocked(invoice.InvoiceID) {
		fmt.Println("locked")
		cancel()
		return
	}

	defer func() {
		log.Println("BYE")
		cancel()
		s.locker.Unlock(invoice.InvoiceID)
	}()

	s.locker.Lock(invoice.InvoiceID)

	for {
		select {
		case <-ctx.Done():
			log.Println("canceled")
			return
		default:

			invoiceCache, err := utils.SafeCast[*domain.Invoices](s.cache.Load(invoice.InvoiceID))
			if err != nil {
				if errors.Is(err, utils.ErrNilParam) {
					invoiceDB, err := s.FindAndSaveToCache(invoice.InvoiceID)
					if err != nil {
						s.l.TemplInvoiceErr("find invoice and save error: "+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
						time.Sleep(reconnectDelay)
						continue
					}
					invoice = invoiceDB
					continue
				}

				s.l.TemplInvoiceErr("cast error:"+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
				time.Sleep(reconnectDelay)
				continue
			}

			invoice = invoiceCache

			if invoice == nil {
				s.l.TemplInvoiceErr("invoice is nil", errid, logger.NA, decimal.Zero, logger.NA, logger.NA, logger.NA, logger.NA)
				time.Sleep(reconnectDelay)
				continue
			}

			if invoice.Status.IsPaid() || invoice.IsInProcessing() {
				fmt.Println("!= NOT PAID: ", invoice.Status)
				return
			}

			if time.Now().Unix() > invoice.EndTimestamp {
				log.Println("END")
				return
			}

			var msg *natsdomain.ResIsPaid

			if s.config.Testing.Enabled {
				msg = &natsdomain.ResIsPaid{
					Amount: invoice.Amount,
					Status: "paid",
					Paid:   true,
				}
				time.Sleep(time.Duration(s.config.Testing.TxConfirmDelay) * time.Second)
			} else {
				// checks whether the client has sent crypto to temp wallet
				msg, err = s.n.ReqIsPaid(invoice, tempWalletAddr)
				if err != nil {
					s.l.TemplInvoiceErr("req is paid error:"+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
					time.Sleep(reconnectDelay)
					continue
				}
			}

			log.Println("check")

			wallet, err := s.wallets.FindByInvoiceID(s.db, invoice.InvoiceID)
			if err != nil {
				fmt.Println(err)
				time.Sleep(reconnectDelay)
				continue
			}

			fmt.Println("MSG STATUS", msg.Status)
			fmt.Println("MSG PAID", msg.Paid)
			fmt.Println("MSG AMOUNT", msg.Amount)
			fmt.Println("WALLET BALLANCE", wallet.Balance)

			// update invoice status
			// oldStatus := invoice.Status
			err = s.db.Transaction(func(tx *gorm.DB) error {
				// if msg.Status != invoice.Status.ToString() {
				// 	invoice.Status = domain.StrToStatus(msg.Status)

				// 	err := s.repo.Update(tx, invoice)
				// 	if err != nil {
				// 		return err
				// 	}

				// 	cache.SaveInvoice(invoice.InvoiceID, invoice)
				// }

				// return fmt.Errorf("test")
				if msg.Amount.GreaterThan(wallet.Balance) {
					fmt.Println("msg.Amount.GreaterThan(wallet.Balance) ", msg.Amount)
					return tx.Model(wallet).Update("balance", gorm.Expr("balance + ?", msg.Amount)).Error
				}

				return nil
			})
			if err != nil {
				s.l.TemplInvoiceErr("global invoice update error: "+err.Error(), invoice.InvoiceID, errid, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
				// cache.SetInvoicePaymentStatus(invoice.InvoiceID, invoice, oldStatus) // rollback to the old status in cache
				time.Sleep(reconnectDelay)
				continue
			}

			if !msg.Paid {
				time.Sleep(reconnectDelay)
				continue
			}

			{ // if paid
				time.Sleep(5 * time.Second) // need to wait for balance update

				invoice.Status = domain.STATUS_IN_PROCESSING
				err := s.repo.Update(s.db, invoice)
				if err != nil {
					fmt.Println(err)
					time.Sleep(reconnectDelay)
					continue
				}
				s.cache.Set(invoice.InvoiceID, invoice, time.Minute*5)

				// withdrawal
				fmt.Println("LETSSS WITHDRAW")
				balance, err := s.balances.Find(s.db, invoice.MerchantID, wallet.Crypto)
				if err != nil {
					fmt.Println(err)
					time.Sleep(reconnectDelay)
					continue
				}

				{

					if s.config.Testing.Enabled {

						// send processing msg

						time.Sleep(3 * time.Second)

						res := natsdomain.ResWithdrawal{
							MerchantId:   invoice.MerchantID,
							InvoiceId:    invoice.InvoiceID,
							Crypto:       wallet.Crypto,
							Address:      balance.Address,
							Amount:       invoice.Amount,
							Status:       msg.Status,
							TxTempWallet: tempWalletAddr,
							TxHash:       gofakeit.BitcoinAddress(),
							TxStatus:     natsdomain.WithdrawalTxStatusProcessing,
						}

						data, err := json.Marshal(res)
						if err != nil {
							s.l.TemplInvoiceErr(err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
							time.Sleep(reconnectDelay)
							continue
						}

						_, err = s.n.Ns.Js.Publish(context.Background(), natsdomain.SubjResWithdrawal.String(), data, jetstream.WithMsgID(natsdomain.NewMsgId(invoice.InvoiceID, natsdomain.MsgActionInfo)))
						if err != nil {
							s.l.TemplInvoiceErr(err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
							time.Sleep(reconnectDelay)
							continue
						}

						// send withdrawal success msg
						s.l.Debug("SLEEP TX FIN PROCESSING")

						time.Sleep(time.Duration(s.config.Testing.TxFinProcessingDelay) * time.Second)

						s.l.Debug("NO SLEEP")

						res.Status = natsdomain.WithdrawalTxStatusSent

						data, err = json.Marshal(natsdomain.ResWithdrawal{
							MerchantId:   invoice.MerchantID,
							InvoiceId:    invoice.InvoiceID,
							Crypto:       wallet.Crypto,
							Address:      balance.Address,
							Amount:       invoice.Amount,
							Status:       msg.Status,
							TxTempWallet: tempWalletAddr,
							TxHash:       gofakeit.BitcoinAddress(),
							TxStatus:     natsdomain.WithdrawalTxStatusSent,
						})

						if err != nil {
							s.l.TemplInvoiceErr(err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
							time.Sleep(reconnectDelay)
							continue
						}

						s.n.Ns.Js.Publish(context.Background(), natsdomain.SubjResWithdrawal.String(), data, jetstream.WithMsgID(natsdomain.NewMsgId(invoice.InvoiceID, natsdomain.MsgActionSuccess)))
						return
					}

				}

				err = s.n.ReqWithdrawal(invoice, wallet, balance, msg.Status)
				if err != nil {
					s.l.TemplInvoiceErr(err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
					time.Sleep(reconnectDelay)
					continue
				}

			}

			log.Println("check 3")

			time.Sleep(10 * time.Second)

		}
	}

}

func (s *InvoicesService) CalculateFinAmount(amount, rate decimal.Decimal, ceil ...int32) decimal.Decimal {
	var _ceil int32

	if ceil != nil && len(ceil) >= 0 {
		_ceil = ceil[0]
	} else { // default ceil
		_ceil = 10
	}

	return amount.Div(rate).RoundCeil(_ceil)
}

func (s *InvoicesService) RunFindEnd() {

	var invoices []domain.Invoices
	s.db.Where(&domain.Invoices{Status: domain.STATUS_IN_PROCESSING}).Find(&invoices)

	for _, i := range invoices {
		if time.Now().Unix() > i.EndTimestamp && !i.IsInProcessing() && !i.Status.IsPaid() {
			i.Status = domain.STATUS_END
			s.db.Save(&i)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *InvoicesService) RunAutostartCheck() {

	var invoices []*domain.Invoices
	s.db.Where(&domain.Invoices{Status: domain.STATUS_NOT_PAID}).Find(&invoices)

	for c, i := range invoices {
		if time.Now().Unix() > i.EndTimestamp && !i.IsInProcessing() {
			continue
		}

		wallet, err := s.wallets.FindByInvoiceID(s.db, i.InvoiceID)
		if err != nil {
			if !postgres.IsNotFound(err) {
				fmt.Println(err)
			}
			continue
		}

		s.cache.Set(i.InvoiceID, i, time.Minute*5)

		ctx, cancel := context.WithTimeout(context.Background(), time.Until(time.Unix(i.EndTimestamp, 0)))
		go s.RunCheck(ctx, cancel, i, wallet.Address)
		fmt.Println("COUNT:", c)
		time.Sleep(500 * time.Millisecond)

	}
}
