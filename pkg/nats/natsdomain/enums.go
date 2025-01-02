package natsdomain

type MerchantWithdrawalStatus uint8

var MerchantWithdrawalTypes = [...]string{"sent", "error"}

const (
	MerchantWithdrawalStatusSent MerchantWithdrawalStatus = iota
	MerchantWithdrawalStatusError
)

type SearchByType uint8

var SearchTypes = [...]string{"hash", "address"}

const (
	SearchByHash SearchByType = iota
	SearchByAddress
)

func (s SearchByType) String() string {
	return SearchTypes[s]
}

type ActionType string

const (
	// blockchain -> api
	MsgActionError ActionType = "error"
	// blockchain -> api
	MsgActionInfo ActionType = "info"
	// blockchain -> api
	MsgActionSuccess ActionType = "success"
	// api -> blockchain
	MsgActionWithdrawal ActionType = "withdrawal"

	// api -> blockchain
	MsgActionWithdrawalRetry ActionType = "withdrawal_retry"
)

// subjects for nats

var KvBuckets = [...]string{"withdraw_status"}

// .js. - jetstream
var SubjectsJetStream = [...]string{"currencies.js.withdraw", "currencies.js.merchant_withdrawal"}

// .core. - nats core
var Subjects = [...]string{"currencies.core.get_tx_status", "currencies.core.new_wallet", "currencies.core.is_paid", "currencies.core.get_rates", "currencies.core.ping", "currencies.core.get_balance"}

var ResponseSubjects = [...]string{"response.withdrawal", "response.merchant-withdrawal"}

type SubjType uint8
type SubjJsType uint8
type SubjResType uint8
type BucketType uint8

// nats core subjects
const (
	// currencies.get_balance
	SubjGetTxStatus SubjType = iota
	SubjNewWallet
	SubjIsPayed
	SubjGetRates
	SubjPing
	SubjGetBalance
)

// nats jetstream subjects
const (
	SubjJsWithdraw SubjJsType = iota
	SubjJsMerchantWithdrawal
)

// nats response subjects
const (
	SubjResWithdrawal SubjResType = iota
	SubjResMerchantWithdrawal
)

func (b BucketType) String() string {
	return KvBuckets[b]
}

func (s SubjType) String() string {
	return Subjects[s]
}

func (s SubjJsType) String() string {
	return SubjectsJetStream[s]
}

func (s SubjResType) String() string {
	return ResponseSubjects[s]
}
