package natsdomain

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/shopspring/decimal"
)

// nats struct
type Ns struct {
	Nc *nats.Conn
	Js jetstream.JetStream
}

type Nc struct {
	*nats.Conn
}

type ReqGetTxStatus struct {
	TxHash         string // only for SearchByHash
	TxTempWallet   string
	BalanceAddress string       // this is the address where the money is sent after processing (from tx_temp_wallet) only for SearchByAddress
	SearchBy       SearchByType // search by (hash/address)
	Cryptocurrency string
}

type ResGetTxStatus struct {
	TxHash    string
	IsPending bool
	Success   bool
	Amount    decimal.Decimal
	Balance   decimal.Decimal
}

type ReqGetBalance struct {
	Cryptocurrency string
	Address        string
}

type ResGetBalance struct {
	Cryptocurrency string
	Balance        decimal.Decimal
}

type Error struct {
	IsError   bool
	Message   string
	Timestamp time.Time
}

type ReqIsPaid struct {
	Address        string
	Cryptocurrency string
	TargetAmount   decimal.Decimal // final amount
}

type ResIsPaid struct {
	Paid   bool
	Status string // not_paid / paid / paid_less
	Amount decimal.Decimal
}

type Withdrawal struct {
	FromAddress string
	MerchantId  string
	ToAddress   string
	Private     string          // from (private key)
	Crypto      string          //cryptocurrency
	Amount      decimal.Decimal // final amount
	// Status     string          // not_paid / paid / paid_less
	// Commission float64
}

type ReqMerchantWithdrawal struct {
	Withdrawal
	UserId              string
	WithdrawalTimestamp string
	// Comission float64 // in % (1.5)
}

type ResMerchantWithdrawal struct {
	Error struct {
		Msg string
	}
	Status MerchantWithdrawalStatus // sent | withdrawal_error
	TxHash string
	ReqMerchantWithdrawal
}

type ReqWithdrawal struct {
	MerchantId   string
	InvoiceId    string
	Address      string
	Private      string          // from (private key)
	Crypto       string          //cryptocurrency
	Amount       decimal.Decimal // final amount
	Status       string          // not_paid / paid / paid_less
	TxTempWallet string          // from this address send money to Address (needed for outbox)
	// Comission float64 not used
}

type ReqNewWallet struct {
	Cryptocurrency string
}

type ReqGetRates struct {
	ToCurrency string // USD/EUR/RUB
}

type Rates struct {
	Eth        decimal.Decimal
	Ltc        decimal.Decimal
	Sol        decimal.Decimal
	Ton        decimal.Decimal
	ToCurrency string
}

type ResGetRates struct {
	Error
	Rates Rates
}

type ResNewWallet struct {
	Address    string
	PrivateKey string
}

const (
	WithdrawalTxStatusProcessing = "processing"
	WithdrawalTxStatusError      = "error"
	WithdrawalTxStatusSent       = "sent"
)

type ResWithdrawal struct {
	Error struct {
		Msg string
	}
	MerchantId string
	InvoiceId  string
	Crypto     string
	Address    string
	Amount     decimal.Decimal
	// Balance    decimal.Decimal
	Status       string
	TxTempWallet string
	TxHash       string
	TxStatus     string
}

// when blockchain service withdraws money
type ResBalanceActions struct {
	Action  string // add / withdraw
	Private string
	Amount  decimal.Decimal
}

// type KvWithdrawStatus struct {
// 	InvoiceId string
// 	Status    string // success / ok / error
// 	TxHash    string
// }
