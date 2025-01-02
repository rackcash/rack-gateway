package sol

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/shopspring/decimal"
)

func NewWallet() (address string, privateKey string, err error) {
	priv, err := solana.NewRandomPrivateKey()
	if err != nil {
		return "", "", err
	}
	address = priv.PublicKey().String()
	privateKey = priv.String()

	return address, privateKey, nil
}

func getBalance(ctx context.Context, client *rpc.Client, address string) (uint64, error) {
	fmt.Println("GET BALANCE START")

	pbc, err := StringToPBC(address)
	if err != nil {
		return 0, err
	}

	balance, err := client.GetBalance(ctx, pbc, rpc.CommitmentFinalized)
	if err != nil {
		// (*jsonrpc.RPCError)(0xc0029854a0)({
		// 	Code: (int) 429,
		// 	Message: (string) (len=82) "Too many requests from your IP, contact your app developer or support@rpcpool.com.",
		// 	Data: (interface {}) <nil>
		//    })

		return 0, err
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("ctx done")
			fmt.Println("GET BALANCE END DONE: ", balance.Value)
			return balance.Value, nil
		default:
			fmt.Println("default fuck")
			fmt.Println("GET BALANCE END: ", balance.Value)

			return balance.Value, nil

		}
	}

	fmt.Println("GET BALANCE END: ", balance.Value)

	return balance.Value, nil

}

func GetBalanceSol(ctx context.Context, client *rpc.Client, address string) (decimal.Decimal, error) {
	balanceLam, err := getBalance(ctx, client, address)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return LamportsToSol(int64(balanceLam)), nil
}

func GetBalanceLamports(ctx context.Context, client *rpc.Client, address string) (uint64, error) {
	return getBalance(ctx, client, address)
}

func StringToPBC(addr string) (solana.PublicKey, error) {
	return solana.PublicKeyFromBase58(addr)
}

func StringToPriv(privateKey string) (solana.PrivateKey, error) {
	return solana.PrivateKeyFromBase58(privateKey)
}

func LamportsToSol(lamports int64) decimal.Decimal {
	const LAMPORTS_PER_SOL = 1e9

	lamportsDecimal := decimal.NewFromInt(int64(lamports))

	sol := lamportsDecimal.Div(decimal.NewFromFloat(LAMPORTS_PER_SOL))

	sol = sol.Round(5)

	return sol
}

func CreateTx(client *rpc.Client, amountLamports uint64, from solana.PrivateKey, to solana.PublicKey) (*solana.Transaction, string, error) {

	if amountLamports <= 0 {
		return nil, "", fmt.Errorf("amountLamports <= 0")
	}

	recent, err := client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, "", err
	}

	fee := uint64(SOL_COMISSION)
	finalAmount := amountLamports - fee

	fmt.Println("AMOUNT: ", amountLamports)
	fmt.Println("FINAL AMOUNT: ", finalAmount)

	if finalAmount <= 0 {
		return nil, "", fmt.Errorf("insufficient funds: %d (amount) <= %d (fee)", amountLamports, fee)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				finalAmount,
				from.PublicKey(),
				to,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	if err != nil {
		return nil, "", err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if from.PublicKey().Equals(key) {
				return &from
			}
			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}

	// tx.Signatures[0].String() - tx hash
	if tx.Signatures == nil {
		return nil, "", fmt.Errorf("tx.Signatures == nil")
	}

	return tx, tx.Signatures[0].String(), nil
}

func SendTx(client *rpc.Client, wsAtomic *atomic.Pointer[ws.Client], tx *solana.Transaction) (solana.Signature, error) {

	err := UpdateWs(wsAtomic)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("ws error: %w", err)
	}

	wsPtr := wsAtomic.Load()
	if wsPtr == nil {
		return solana.Signature{}, fmt.Errorf("ws is nil")
	}

	sig, err := confirm.SendAndConfirmTransaction(
		context.Background(),
		client,
		wsPtr,
		tx,
	)

	if err != nil {
		return solana.Signature{}, err
	}

	return sig, nil
}

func GetTxByHash(client *rpc.Client, hash string) (amount decimal.Decimal, success bool, err error) {

	txSig, err := solana.SignatureFromBase58(hash)
	if err != nil {
		return decimal.Decimal{}, false, err
	}

	tx, err := client.GetTransaction(
		context.Background(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return decimal.Decimal{}, false, err
	}

	if tx.Meta != nil {
		senderIndex := 0
		// recipientIndex := 1

		// Получаем балансы отправителя и получателя до и после транзакции
		senderPreBalance := tx.Meta.PreBalances[senderIndex]
		senderPostBalance := tx.Meta.PostBalances[senderIndex]
		// recipientPreBalance := tx.Meta.PreBalances[recipientIndex]
		// recipientPostBalance := tx.Meta.PostBalances[recipientIndex]

		// // Вычисляем сумму перевода
		// fmt.Println(senderPreBalance, senderPostBalance)
		// amount := senderPreBalance - senderPostBalance
		// fmt.Printf("Transaction Amount: %v\n", LamportsToSol(int64(amount)))
		// fmt.Println(recipientPreBalance, recipientPostBalance)

		fmt.Printf("SENDER PRE BALANCE: %v\n", LamportsToSol(int64(senderPreBalance)))
		fmt.Printf("SENDER POST BALANCE: %v\n", LamportsToSol(int64(senderPostBalance)))
		fmt.Println("AMOUNT SOL: ", LamportsToSol(int64(senderPreBalance-SOL_COMISSION)))

		/* нужно отнять комиссию (sol.CreateTx -> fee) */

		return LamportsToSol(int64(senderPreBalance - SOL_COMISSION)), true, nil

	}

	return decimal.Decimal{}, false, fmt.Errorf("tx.Meta == nil")

}
