package service

import (
	"encoding/json"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/nats"
	"infra/api/internal/logger"
	"infra/api/internal/repository"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type OutboxEventsService struct {
	repo                 repository.Events
	invoicesService      Invoices
	getWithdrawalService GetWithdrawal
	webhook              WebhookSender

	balances  repository.Balances
	wallets   repository.Wallets
	natsinfra *nats.NatsInfra

	db *gorm.DB
	l  logger.Logger
}

func NewOutboxEventsService(invoicesService Invoices, balances repository.Balances, wallets repository.Wallets, getWithdrawalService GetWithdrawal, natsinfra *nats.NatsInfra, db *gorm.DB, l logger.Logger, repo repository.Events, webhook WebhookSender) *OutboxEventsService {
	return &OutboxEventsService{invoicesService: invoicesService, balances: balances, wallets: wallets, natsinfra: natsinfra, db: db, l: l, getWithdrawalService: getWithdrawalService, repo: repo, webhook: webhook}
}

// checks events table and handles them
func (s *OutboxEventsService) StartProcessEvents() {
	const sleepTime = 10 * time.Second
	fmt.Println("OUTBOX EVENTS START")

	go func() {
		for {
			events, err := getNewEvent(s.db, 20, time.Second*1, s.l)
			if err != nil {
				fmt.Println(err)
				time.Sleep(sleepTime)
				continue
			}

			for _, event := range events {
				switch event.Type {
				case domain.EVENT_INVOICE_PROCESSING:
					s.handleInvoiceProcessingEvent(event)
				case domain.EVENT_WEBHOOK:
					s.handleWebhookEvent(event)
				default:
					fmt.Println("Invalid event type: " + event.Type)
					continue
				}

			}

			fmt.Println("EVENTS", events)

			time.Sleep(sleepTime)
		}
	}()

}

func (s *OutboxEventsService) handleWebhookEvent(event domain.Events) {
	payload, err := utils.Unmarshal[domain.WebhookPayload]([]byte(event.Payload))
	if err != nil {
		fmt.Println("Unmarshal[domain.WebhookPayload]: ", err)
		return
	}

	go func() {
		if err := s.webhook.Send(payload.Url, payload.Info); err != nil {
			s.l.Debug("send webhook error: "+err.Error(), "url", payload.Url, "info", payload.Info)
		}
		s.repo.Done(s.db, event.RelationID, domain.EVENT_WEBHOOK)
	}()
}

func (s *OutboxEventsService) handleInvoiceProcessingEvent(event domain.Events) {
	// const timeDelta = time.Hour * 5
	// is, s := helpers.IsTimeDeltaExceeded(event.CreatedAt, timeDelta)
	// fmt.Println("SINCE", s)
	// if !is {
	// 	fmt.Println("Not time delta exceeded")
	// }

	payload, err := utils.Unmarshal[domain.PayloadInvoiceProcessing]([]byte(event.Payload))
	if err != nil {
		fmt.Println("Unmarshal[domain.PayloadInvoiceProcessing]: ", err)
		return
	}
	fmt.Println(payload)

	invoice, err := s.invoicesService.FindByID(s.db, payload.InvoiceID)
	if err != nil {
		fmt.Println(err)
		return
	}

	wallet, err := s.wallets.FindByInvoiceID(s.db, invoice.InvoiceID)
	if err != nil {
		s.l.Debug(err.Error())
		return
	}

	balance, err := s.balances.Find(s.db, invoice.MerchantID, invoice.Cryptocurrency)
	if err != nil {
		s.l.Debug(err.Error())
		return
	}

	var reqGetStatus = natsdomain.ReqGetTxStatus{Cryptocurrency: invoice.Cryptocurrency, TxTempWallet: wallet.Address}
	//  TODO: add more cryptocurrencies
	switch invoice.Cryptocurrency {
	case "eth":
		reqGetStatus.SearchBy = natsdomain.SearchByHash
		reqGetStatus.TxHash = payload.TxHash
	case "sol":
		reqGetStatus.SearchBy = natsdomain.SearchByHash
		reqGetStatus.TxHash = payload.TxHash
	case "ton":
		reqGetStatus.SearchBy = natsdomain.SearchByAddress
		reqGetStatus.BalanceAddress = balance.Address
	default:
		fmt.Println("Invalid cryptocurrency")
	}

	jsonReq, err := json.Marshal(reqGetStatus)
	if err != nil {
		fmt.Println(err)
		return
	}

	msg, err := s.natsinfra.ReqAndRecv(natsdomain.SubjGetTxStatus, jsonReq)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("MSG", string(msg))

	is, errmsg := nats.HelpersIsError(msg)
	if is {
		s.l.Debug("n.HelpersIsError(msg): "+errmsg, "tx hash", payload.TxHash)
		return
	}

	jsonMsg, err := utils.Unmarshal[natsdomain.ResGetTxStatus](msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	if jsonMsg.TxHash != payload.TxHash {
		s.l.Debug("jsonMsg.TxHash != payload.TxHash")
		return
	}

	if jsonMsg.Success && jsonMsg.Amount.GreaterThan(decimal.Zero) { // sent to the main wallet
		convertedStatus := domain.StrToStatus(payload.Status)

		s.l.Debug("Payload status", "String", payload.Status, "Int8", convertedStatus)

		err = s.db.Transaction(func(tx *gorm.DB) error {
			// FIXME: uncomment
			// err := setPaidStatus(app, tx, balance, invoice, jsonMsg.Amount, convertedStatus)
			return s.getWithdrawalService.SetPaidStatus(tx, balance, invoice, convertedStatus)
		})

		s.l.Debug("tx status", "error", err)

		return
	}

	if jsonMsg.IsPending {
		s.l.Debug("Pending")
		return
	}

	s.l.Debug("Not sent to the main wallet")

	if balance.Balance.Equal(decimal.Zero) {
		s.l.Debug("balance.Balance.Equal(helpers.ZERO_DECIMAL)")
		return
	}

	err = s.natsinfra.ReqWithdrawal(invoice, wallet, balance, payload.Status)
	if err != nil {
		s.l.Debug(err.Error())
		return
	}

	fmt.Println("jsonMsg", jsonMsg)

}

func selectEventsFromDb(tx *gorm.DB, count int) ([]domain.Events, error) {
	var events []domain.Events
	return events, tx.Where(&domain.Events{Status: "new"}).Limit(count).Find(&events).Error
}

const errNoValidEvents = "no valid events"

func getNewEvent(tx *gorm.DB, count int, timeDiff time.Duration, log logger.Logger) ([]domain.Events, error) {
	var validEvents []domain.Events

	fmt.Println("getNewEvent")
	events, err := selectEventsFromDb(tx, count)
	if err != nil {
		return nil, err
	}
	// fmt.Println(events)

	// filter events by time

	for _, x := range events {
		duration := time.Since(x.CreatedAt)
		fmt.Println(duration > timeDiff)
		if duration > timeDiff {
			validEvents = append(validEvents, x)
		}

	}

	if len(validEvents) == 0 {
		return nil, fmt.Errorf(errNoValidEvents)
	}

	return validEvents, nil

}
