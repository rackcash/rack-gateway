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
	"infra/pkg/utils"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GetMerchantWithdrawalService struct {
	// repo repository
	// invoicesService  Invoices
	// merchantsService Merchants
	balances    repository.Balances
	withdrawals repository.Withdrawals
	// wallets          repository.Wallets
	// webhook          WebhookSender
	// events           repository.Events

	config    *config.Config
	c         jetstream.Consumer
	l         logger.Logger
	db        *gorm.DB
	natsinfra *nats.NatsInfra
}

func NewGetMerchantWithdrawalService(db *gorm.DB, natsinfra *nats.NatsInfra, l logger.Logger, balances repository.Balances, withdrawals repository.Withdrawals, config *config.Config) *GetMerchantWithdrawalService {
	stream, err := nats.InitResponsesStream(context.Background(), natsinfra.Js)
	if err != nil {
		panic(err)
	}

	c, err := stream.CreateOrUpdateConsumer(context.Background(), jetstream.ConsumerConfig{
		Durable:       "merchant_withdrawal_status",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: natsdomain.SubjResMerchantWithdrawal.String(),
	})
	if err != nil {
		panic("CreateOrUpdateConsumer error" + err.Error())
	}

	return &GetMerchantWithdrawalService{db: db, natsinfra: natsinfra, c: c, balances: balances, l: l, config: config, withdrawals: withdrawals}
}

func (s *GetMerchantWithdrawalService) StartWaitStatus() {

	const delay = time.Second * 10

	_, err := s.c.Consume(func(msg jetstream.Msg) {
		err := s.consumer(msg)
		if err != nil {
			s.l.Debug("Consume error", "error", err.Error(), "msg", string(msg.Data()))
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

func (s *GetMerchantWithdrawalService) consumer(msg jetstream.Msg) error {

	fmt.Println("Received a message", string(msg.Data()))

	m, _ := msg.Metadata()
	if m != nil {
		if m.NumDelivered > 3 {
			s.l.Debug("Too many deliveries", "num", m.NumDelivered)
			return nil
		}
	}

	jsonRes, err := utils.Unmarshal[natsdomain.ResMerchantWithdrawal](msg.Data())
	if err != nil {
		fmt.Println("Unmarshal error", err)
		return err
	}

	if jsonRes.MerchantId == "" {
		fmt.Println("Withdraw error")
		return fmt.Errorf("withdraw error: jsonRes.MerchantId is empty")
	}

	// merchant, err := s.merchantsService.FindByID(s.db, jsonRes.MerchantId)
	// if err != nil {
	// 	fmt.Println("Find merchant by id error", err)
	// 	return err
	// }

	switch jsonRes.Status {
	case natsdomain.MerchantWithdrawalStatusSent:
		return s.handleSent(jsonRes)
	case natsdomain.MerchantWithdrawalStatusError:
		fmt.Println("error")
		return s.handleError(jsonRes)
	default:
		s.l.Debug("Invalid status", "status", jsonRes.Status)
	}

	return nil

}

func (s *GetMerchantWithdrawalService) handleError(jsonRes *natsdomain.ResMerchantWithdrawal) error {
	var errid = logger.GenErrorId()

	data, err := json.Marshal(natsdomain.ReqMerchantWithdrawal{
		WithdrawalTimestamp: jsonRes.WithdrawalTimestamp,
		WithdrawalID:        jsonRes.WithdrawalID,
		Withdrawal: natsdomain.Withdrawal{
			FromAddress: jsonRes.FromAddress,
			MerchantId:  jsonRes.MerchantId,
			ToAddress:   jsonRes.ToAddress,
			Private:     jsonRes.Private,
			Crypto:      jsonRes.Crypto,
			Amount:      jsonRes.Amount,
		},
	})
	if err != nil {
		return err
	}

	// TODO: change withdrawal status to error
	err = s.withdrawals.UpdateStatus(s.db, jsonRes.WithdrawalID, domain.WITHDRAWAL_ERROR)
	if err != nil {
		s.l.TemplInvoiceErr("update withdrawal status error: "+err.Error(), errid, jsonRes.WithdrawalTimestamp, jsonRes.Amount, jsonRes.Crypto, logger.NA, jsonRes.MerchantId, logger.NA)
	}

	return s.natsinfra.JsPublishMsgId(natsdomain.SubjJsMerchantWithdrawal.String(), data, natsdomain.NewMsgId(jsonRes.WithdrawalTimestamp+jsonRes.MerchantId, natsdomain.MsgActionWithdrawalRetry))
}

func (s *GetMerchantWithdrawalService) handleSent(jsonRes *natsdomain.ResMerchantWithdrawal) error {
	var errid = logger.GenErrorId()

	// update balance in db

	prevBalance, err := s.balances.Find(s.db, jsonRes.MerchantId, jsonRes.Crypto)
	if err != nil {
		return err
	}

	fmt.Println("FROM ADDRESS", jsonRes.FromAddress)
	newBalance, err := s.getBalance(jsonRes.MerchantId, prevBalance.Balance, jsonRes.Crypto, jsonRes.FromAddress)
	if err != nil {
		return err
	}

	s.l.Debug("Update balance")

	s.db.Transaction(func(tx *gorm.DB) error {
		err = s.withdrawals.UpdateStatus(tx, jsonRes.WithdrawalID, domain.WITHDRAWAL_SUCCESS)
		if err != nil {
			s.l.TemplInvoiceErr("update withdrawal status error: "+err.Error(), errid, jsonRes.WithdrawalTimestamp, jsonRes.Amount, jsonRes.Crypto, logger.NA, jsonRes.MerchantId, logger.NA)
			return err
		}

		if err := tx.Model(prevBalance).Updates(domain.Balances{Balance: newBalance.Balance}).Error; err != nil {
			return err
		}
		return nil

	})

	// TODO: refactoring
	// send notification to telegram
	// err = func() error {
	// 	const retries = 5
	// 	const delay = time.Second * 30

	// 	for i := 0; i < retries; i++ {
	// 		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.config.
	// 			Telegram.Token)

	// 		text := fmt.Sprintf("💰 %s %s sent to your address (%s)", jsonRes.Amount, jsonRes.Crypto, jsonRes.ToAddress)

	// 		jsonData := fmt.Sprintf(`{
	// 	"chat_id": "%s",
	// 	"text": "%s"
	// }`, jsonRes.UserId, text)

	// 		s.l.Debug("Send notification to telegram", "url", url, "jsonData", jsonData)

	// 		resp, err := http.Post(url, "application/json", bytes.NewBufferString(jsonData))
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defer resp.Body.Close()

	// 		if resp.StatusCode == 200 { // OK
	// 			return nil
	// 		}

	// 		if resp.StatusCode == 429 { // Too Many Requests
	// 			time.Sleep(delay)
	// 			continue
	// 		}

	// 		if resp.StatusCode >= 500 { // internal server error
	// 			time.Sleep(delay)
	// 			continue
	// 		}

	// 		// other status code
	// 		return fmt.Errorf("http post returned an unknown status code %d", resp.StatusCode)
	// 	}
	// 	return nil
	// }()

	return err
}

func (s *GetMerchantWithdrawalService) getBalance(merchantId string, prevBalance decimal.Decimal, crypto, address string) (*natsdomain.ResGetBalance, error) {

	s.l.Debug("Get Balance", "address", address)
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

	b, err = s.natsinfra.ReqGetBalance(crypto, address)
	if err != nil {
		time.Sleep(delay)
		goto start
	}

	// BUG; nats balance when 0 shows 0.025814454365697
	s.l.Debug("Balances", "merchant id", merchantId, "db balance", prevBalance, "nats balance", b.Balance, "attempts", attempts)

	if b.Balance.GreaterThanOrEqual(prevBalance) { // prevBalance должен быть меньше
		fmt.Println("GTE", b.Balance, prevBalance)
		err = fmt.Errorf(STATUS_NOT_CHANGED)
		time.Sleep(delay)
		goto start
	}

	return b, nil
}
