package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"infra/api/internal/config"
	"infra/api/internal/domain"
	"infra/api/internal/infra/nats"
	"infra/api/internal/logger"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/pgerror"
	"infra/pkg/utils"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nats-io/nats.go/jetstream"
	"gorm.io/gorm"
)

type GetWithdrawalService struct {
	// repo repository
	invoicesService Invoices
	balances        repository.Balances
	wallets         repository.Wallets
	webhook         WebhookSender
	events          repository.Events

	config    *config.Config
	c         jetstream.Consumer
	l         logger.Logger
	db        *gorm.DB
	natsinfra *nats.NatsInfra
}

func NewGetWithdrawalService(db *gorm.DB, natsinfra *nats.NatsInfra, l logger.Logger, events repository.Events, wallets repository.Wallets, balances repository.Balances, invoicesService Invoices, webhookSender WebhookSender, config *config.Config) *GetWithdrawalService {
	stream, err := nats.InitResponsesStream(context.Background(), natsinfra.Js)
	if err != nil {
		panic(err)
	}

	c, err := stream.CreateOrUpdateConsumer(context.Background(), jetstream.ConsumerConfig{
		Durable:       "withdrawal_status",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: natsdomain.SubjResWithdrawal.String(),
	})
	if err != nil {
		panic("CreateOrUpdateConsumer error" + err.Error())
	}

	return &GetWithdrawalService{db: db, natsinfra: natsinfra, c: c, invoicesService: invoicesService, balances: balances, wallets: wallets, events: events, l: l, webhook: webhookSender, config: config}
}

func (s *GetWithdrawalService) StartWaitStatus() {
	const delay = time.Second * 10

	_, err := s.c.Consume(func(msg jetstream.Msg) {
		err := s.consumer(msg)
		if err != nil {
			msg.NakWithDelay(delay)
			return
		}
		fmt.Println(msg.Ack())
	})

	if err != nil {
		s.l.TemplNatsError("Consume error", s.natsinfra.Nc.ConnectedUrl(), err)
		return
	}

}

func (s *GetWithdrawalService) consumer(msg jetstream.Msg) error {
	// if msg.Subject() == domain.SubjResWithdrawal.String() {
	// 	fmt.Println("Invalid subject: " + msg.Subject())
	// 	return
	// }

	fmt.Println("Received a message", string(msg.Data()))

	m, _ := msg.Metadata()
	if m != nil {
		if m.NumDelivered > 3 {
			s.l.Debug("Too many deliveries", "num", m.NumDelivered)
			return nil
		}
	}

	if string(msg.Data()) == "test" {
		// return fmt.Errorf("helloo")
		return nil
	}

	jsonRes, err := utils.Unmarshal[natsdomain.ResWithdrawal](msg.Data())
	if err != nil {
		fmt.Println("Unmarshal error", err)
		return err
	}

	if jsonRes.InvoiceId == "" {
		fmt.Println("Withdraw error")
		return fmt.Errorf("withdraw error")
	}

	invoice, err := s.invoicesService.FindByID(s.db, jsonRes.InvoiceId)
	if err != nil {
		fmt.Println("FindInvoiceByID error", err)
		return err
	}

	// handling

	switch jsonRes.TxStatus {
	case natsdomain.WithdrawalTxStatusError:
		return s.handleTxError(jsonRes, invoice)
	case natsdomain.WithdrawalTxStatusProcessing:
		return s.handleProcessing(jsonRes, invoice)
	case natsdomain.WithdrawalTxStatusSent:
		fmt.Println("paid")
		return s.handlePaid(jsonRes, invoice)
	default:
		s.l.Debug("Invalid tx status", "tx status", jsonRes.TxStatus)

	}

	fmt.Println("ENNDDD")

	return nil
}

func (s *GetWithdrawalService) handlePaid(jsonRes *natsdomain.ResWithdrawal, invoice *domain.Invoices) error {

	var errid = logger.GenErrorId()
	invoice.Status = domain.StrToStatus(jsonRes.Status)

	balance, err := s.balances.Find(s.db, invoice.MerchantID, invoice.Cryptocurrency)
	if err != nil {
		s.l.TemplInvoiceErr("find balance error: "+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
		return err
	}

	s.l.Debug("Got withdraw response", "status", jsonRes.Status, "tx status", jsonRes.TxStatus)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		err := s.SetPaidStatus(tx, balance, invoice, invoice.Status)
		if err != nil {
			fmt.Println("INVOICE PAYMENT STATUS", invoice.Status)
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Println("ERROROR", err)
		return err
	}

	// TODO: webhook http req

	fmt.Println("WEBHOOK URL", invoice.Webhook)
	if invoice.Webhook != "" {

		event, err := s.events.Find(s.db, invoice.ID, domain.EVENT_WEBHOOK)
		if err != nil {
			fmt.Println("FIND EVENT ERROR", err)
		}

		if event != nil {
			fmt.Println("EVENT: ", event.ID)
		}

		time.Sleep(time.Second * 5)

		var response = domain.ResponseInvoiceInfo{
			Id:     invoice.InvoiceID,
			Amount: invoice.Amount.String(),
			// Currency:       invoice.Cryptocurrency,
			// CryptoAmount:   invoice.CryptoAmount.String(),
			Cryptocurrency: invoice.Cryptocurrency,
			IsPaid:         invoice.Status.IsPaid(),
			Status:         invoice.Status.ToString(),
			CreatedAt:      invoice.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if time.Now().Unix() > invoice.EndTimestamp && invoice.Status.IsNotPaid() {
			response.Status = "end"
		}

		// create outbox event

		eventPayload, err := json.Marshal(domain.WebhookPayload{
			MerchantID: invoice.MerchantID,
			InvoiceID:  invoice.InvoiceID,
			Url:        invoice.Webhook,
			Info:       response,
		})
		if err != nil {
			s.l.Debug(err.Error())
			return err
		}

		fmt.Println(string(eventPayload))

		if err := s.events.Create(s.db, domain.EVENT_WEBHOOK, invoice.ID, string(eventPayload)); err != nil {
			s.l.Debug(err.Error())
			return err
		}

		// send webhook

		// TODO:  test

		s.l.Debug("SLEEP BEFORE WEBHOOk")
		time.Sleep(time.Second * 5)

		fmt.Println("SENDING WEBHOOK")

		err = s.webhook.Send(invoice.Webhook, response)
		if err != nil {
			s.l.Debug(err.Error())
			// return
		}

		// finish
		s.events.Done(s.db, invoice.ID, domain.EVENT_WEBHOOK)
	}

	return nil

}

func (s *GetWithdrawalService) handleProcessing(jsonRes *natsdomain.ResWithdrawal, invoice *domain.Invoices) error {
	// eventPayload := fmt.Sprintf(
	// 	`{"invoice_id": "%s", "tx_hash": "%s", "crypto_amount": "%s", "status": "%s"}`,
	// 	jsonRes.InvoiceId,
	// 	jsonRes.TxHash,
	// 	jsonRes.Amount,
	// 	jsonRes.Status,
	// )
	var errid = logger.GenErrorId()

	if !s.config.Testing.Enabled {
		eventPayload := fmt.Sprintf(
			`{"invoice_id": "%s", "tx_hash": "%s", "crypto_amount": "%s", "status": "%s", "tx_temp_wallet": "%s", "balance_address": "%s"}`,
			invoice.InvoiceID,
			jsonRes.TxHash,
			invoice.Cryptocurrency,
			jsonRes.Status,
			jsonRes.TxTempWallet,
			jsonRes.Address, // balance address
		)

		// fmt.Println("EVENT PAYLOAD", eventPayload)

		err := s.events.Create(s.db, domain.EVENT_INVOICE_PROCESSING, invoice.ID, eventPayload)
		if err != nil {
			fmt.Println("EventCreate error", err)

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerror.UniqueViolation { // means outbox handler sent the same event
				return nil
			}

			s.l.TemplInvoiceErr("create outbox event: "+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
			return err
		}
	}

	invoice.ProcessedTxHash = jsonRes.TxHash
	err := s.invoicesService.Update(s.db, invoice)
	if err != nil {
		s.l.TemplInvoiceErr("update invoice: "+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
		return err
	}

	return nil
}

func (s *GetWithdrawalService) handleTxError(jsonRes *natsdomain.ResWithdrawal, invoice *domain.Invoices) error {

	var errid = logger.GenErrorId()

	wallet, err := s.wallets.FindByInvoiceID(s.db, invoice.InvoiceID)
	if err != nil {
		return nil
	}

	balance, err := s.balances.Find(s.db, invoice.MerchantID, invoice.Cryptocurrency)
	if err != nil {
		s.l.TemplInvoiceErr("find balance error: "+err.Error(), errid, invoice.InvoiceID, invoice.Amount, invoice.Cryptocurrency, logger.NA, invoice.MerchantID, logger.NA)
		return err
	}

	data, err := json.Marshal(natsdomain.ReqWithdrawal{InvoiceId: invoice.InvoiceID, MerchantId: invoice.MerchantID, Address: balance.Address, Private: wallet.Private, Crypto: wallet.Crypto, Amount: wallet.Balance, Status: jsonRes.Status, TxTempWallet: wallet.Address})
	if err != nil {
		return err
	}

	err = s.natsinfra.JsPublish(natsdomain.SubjJsWithdraw.String(), data)
	if err != nil {
		return err
	}

	return nil
}

// updates status, balance, invoice, when invoice is paid
func (s *GetWithdrawalService) SetPaidStatus(tx *gorm.DB, balance *domain.Balances, invoice *domain.Invoices, paymentStatus domain.Status) error {

	var b *natsdomain.ResGetBalance
	var err error

	if !s.config.Testing.Enabled {
		b, err = s.getBalance(balance, invoice)

		if err != nil {
			return err
		}

		fmt.Println("BALANCE", b.Balance)
	} else {
		b = &natsdomain.ResGetBalance{
			Cryptocurrency: invoice.Cryptocurrency,
			Balance:        invoice.Amount,
		}
	}

	err = tx.Model(balance).Update("balance", b.Balance).Error
	if err != nil {
		return err
	}

	invoice.Status = paymentStatus

	err = s.invoicesService.UpdateAndSave(tx, invoice)
	if err != nil {
		return err
	}

	err = s.events.Done(tx, invoice.ID, domain.EVENT_INVOICE_PROCESSING)
	if err != nil {
		return err
	}

	// TODO:  webhook http req

	return nil
}

func (s *GetWithdrawalService) getBalance(balance *domain.Balances, invoice *domain.Invoices) (*natsdomain.ResGetBalance, error) {

	const (
		STATUS_NOT_CHANGED = "not changed"
		maxAttempts        = 3
		delay              = time.Second * 3 // delay between attempts
	)

	var (
		attempts = 0
		err      error
		b        *natsdomain.ResGetBalance
	)

start:
	attempts++

	{ // check attempts
		if attempts > maxAttempts {
			// The balance is not changed, if the money is not credited, will check it on the next invoice, because it could be a false positive and the money is credited
			if errors.Is(err, fmt.Errorf(STATUS_NOT_CHANGED)) {
				return nil, nil
			}

			return nil, fmt.Errorf("balance is not received: " + err.Error())
		}
	}

	b, err = s.natsinfra.ReqGetBalance(balance.Crypto, balance.Address)
	if err != nil {
		time.Sleep(delay)
		goto start
	}

	s.l.Debug("Balances", "invoice id", invoice.ID, "db balance", balance.Balance, "nats balance", b.Balance, "attempts", attempts)

	if b.Balance.LessThanOrEqual(balance.Balance) {
		err = fmt.Errorf(STATUS_NOT_CHANGED)
		time.Sleep(delay)
		goto start
	}

	return b, nil
}
