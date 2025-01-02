package nats

import (
	"infra/blockchain/rack/config"
	"infra/pkg/dlog"
	"infra/pkg/failover"
	"infra/pkg/nats/natsdomain"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/shopspring/decimal"
	"github.com/xssnick/tonutils-go/ton"
)

//	type Error struct {
//		IsError   bool
//		Message   string
//		Timestamp time.Time
//	}

type Ton struct {
	Client  ton.APIClientWrapped
	BlockID atomic.Pointer[ton.BlockIDExt]
}

type Sol struct {
	Client *rpc.Client
	Ws     atomic.Pointer[ws.Client]
}

type App struct {
	// TON
	Ton *Ton

	//  ETHEREUM
	EthClient   *ethclient.Client
	EthFailover failover.Failover

	// Solana
	Sol *Sol

	//
	Config *config.Config
	Ns     *natsdomain.Ns
	C      jetstream.Consumer
	Dlog   dlog.Dlog
}

// type Nc struct {
// 	*nats.Conn
// }
// type MsgIsPayed struct {
// 	Address        string
// 	Cryptocurrency string
// 	TargetAmount   decimal.Decimal // final amount
// }

type Withdraw struct {
	MerchantId string
	Address    string
	Private    string          // from (private key)
	Currency   string          //cryptocurrency
	Amount     decimal.Decimal // final amount
	// Status     string          // not_paid / paid / paid_less
	// Commission float64
}

type ReqMerchantWithdrawal struct {
	Withdraw
	UserId string
	// Comission float64 // in % (1.5)
}

// type MsgServerWithdraw struct {
// 	Withdraw
// 	Status    string
// 	InvoiceId string
// }

// type ResClientWithdraw struct {
// 	Error struct {
// 		Msg string
// 	}
// 	Status string // sent | withdraw_error
// 	TxHash string
// 	ReqClientWithdraw
// }

// type ReqIsPaid struct {
// 	Paid   bool
// 	Status string
// 	Amount decimal.Decimal
// }

// type MsgGetBalance struct {
// 	Cryptocurrency string
// 	Address        string
// }

// type ReqGetBalance struct {
// 	Cryptocurrency string
// 	Balance        decimal.Decimal
// }

// type MsgGetTxStatus struct {
// 	FromAddress    string
// 	TxHash         string
// 	Cryptocurrency string
// }

// type ReqGetTxStatus struct {
// 	TxHash    string
// 	Success   bool
// 	IsPending bool
// 	Amount    decimal.Decimal
// 	Balance   decimal.Decimal
// }

// type MsgNewWallet struct {
// 	Cryptocurrency string
// }
// type ReqNewWallet struct {
// 	Address    string
// 	PrivateKey string
// }
// type MsgGetRates struct {
// 	ToCurrency string // USD/EUR/RUB
// }
// type ReqGetRates struct {
// 	Error
// 	Rates currencies.Rates
// }

// type ReqServerWithdraw struct {
// 	Error struct {
// 		Msg string
// 	}
// 	MerchantId string
// 	InvoiceId  string
// 	Currency   string
// 	Address    string
// 	Amount     decimal.Decimal
// 	// Balance    decimal.Decimal
// 	Status string
// 	TxHash string
// }
