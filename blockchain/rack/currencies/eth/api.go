package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

func NewWallet() (address string, privateKey string) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(privKey)

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	// publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println("PrivKey:", hexutil.Encode(privateKeyBytes))

	address = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Address:", address)

	return address, strings.TrimPrefix(hexutil.Encode(privateKeyBytes), "0x")
}

func CreateTx(client *ethclient.Client, toAddress_, privKey string, amount *big.Int, attemptsLimit int) (*types.Transaction, error) {
	var attempts int

	fmt.Println("TX 1")

CreateTx:
	if attempts > attemptsLimit {
		return nil, fmt.Errorf("failed to create tx after %d attempts", attempts)
	}

	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		slog.Debug(err.Error())
		return nil, err
	}

	fmt.Println("TX 2")

	fromAddress := PrivateToAddress(privKey)

	balance, err := GetBalanceWei(client, fromAddress.Hex())
	if err != nil {
		return nil, err
	}

	if balance.Sign() == 0 {
		attempts++
		slog.Debug("Insufficient funds (0): ", "balance", balance)

		time.Sleep(5 * time.Second)
		goto CreateTx
	}

	fmt.Println("TX 3")

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	fmt.Println("TX 3")

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	fmt.Println("AMOUNT WEI", amount)

	var finalAmount *big.Int

	gasCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	availableBalance := new(big.Int).Sub(balance, gasCost)

	if availableBalance.Sign() <= 0 {
		return nil, fmt.Errorf("insufficient funds: balance %s, gas cost %s (%d)", balance, gasCost, WeiToEther(gasCost))
	}

	if amount.Cmp(availableBalance) > 0 {
		finalAmount = availableBalance
	} else {
		finalAmount = amount
	}

	toAddress := common.HexToAddress(toAddress_)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, finalAmount, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
	fmt.Println(GetBalanceEther(client, fromAddress.Hex()))

	return signedTx, nil
}

func GetTxByHash(client *ethclient.Client, _txHash string, tempWallet string) (tx *types.Transaction, success bool, isPending bool, err error) {
	txHash := common.HexToHash(_txHash)

	tx, isPending, err = client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, false, false, err
	}

	if !isPending {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			return nil, false, false, err
		}

		if receipt.Status == 1 {
			success = true
		} else {
			success = false
		}
	}

	return tx, success, isPending, nil

}

func SendTx(client *ethclient.Client, tx *types.Transaction) (*decimal.Decimal, error) {
	fmt.Println("Waiting for transaction to be mined...")
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return nil, err
	}

	if receipt.Status == 1 {
		fmt.Println("Transaction mined successfully")
	} else {
		var zero = decimal.NewFromInt(0)
		return &zero, fmt.Errorf("transaction failed")
	}

	return WeiToEther(tx.Value()), nil
}

func getBalance(client *ethclient.Client, address string) (*big.Int, error) {
	fmt.Println("ADDRESS", address)
	account := common.HexToAddress(address)

	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("BALANCE", balance)

	// balanceEther := decimal.NewFromBigInt(balance, -18)

	// fmt.Println("BALANCE", balanceEther.BigInt())

	return balance, nil
}

func GetBalanceWei(client *ethclient.Client, address string) (*big.Int, error) {
	return getBalance(client, address)
}

func GetBalanceEther(client *ethclient.Client, address string) (decimal.Decimal, error) {
	balance, err := getBalance(client, address)
	if err != nil {
		return decimal.Decimal{}, err
	}

	fmt.Println("BALANCE WEI", balance)

	ether := decimal.NewFromBigInt(balance, -18)
	return ether, nil
}

func bigIntToDecimal(wei *big.Int) decimal.Decimal {
	weiDecimal := decimal.NewFromBigInt(wei, 0)
	return weiDecimal
}

// 1230000000000000000 to 1.23
func WeiToEther(wei *big.Int) *decimal.Decimal {
	etherConversionFactor := decimal.NewFromInt(1000000000000000000) // 1 Ether = 10^18 Wei
	weiDecimal := bigIntToDecimal(wei)
	ether := weiDecimal.Div(etherConversionFactor)
	return &ether
}

// 1.23 to 123000000000000000
func EtherToWei(ether decimal.Decimal) *big.Int {
	weiConversionFactor := decimal.NewFromInt(1000000000000000000) // 1 Ether = 10^18 Wei
	wei := ether.Mul(weiConversionFactor)
	weiBigInt := new(big.Int)
	weiBigInt.SetString(wei.String(), 10)
	return weiBigInt
}

func PrivateToAddress(private string) common.Address {
	privateKey, err := crypto.HexToECDSA(private)
	if err != nil {
		fmt.Println("crypto.HexToECDSA:", err)
		return common.Address{}
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return common.Address{}
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA)

}
