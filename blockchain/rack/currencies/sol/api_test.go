package sol

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/shopspring/decimal"
)

func TestNewWallet(t *testing.T) {
	var wg sync.WaitGroup
	const N = 10000

	wg.Add(N)

	for range N {
		go func() {
			defer wg.Done()
			_, _, err := NewWallet()
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
}

func TestGetBalance(t *testing.T) {
	tests := []struct {
		address string
		balance decimal.Decimal
	}{
		{"9eVraFo7pidgxsLbPa9QLPyhYEPLA5vgUZ9zZTpDmyDr", decimal.NewFromFloat(0.039985)},
		{"Bt93fJKFwXsQXqKKTZ4rpavrrNPBQzbYzpYYmZqbrqT9", decimal.NewFromFloat(0.01)},
		{"9rVz3dfegDTh5Wf1tD7y4k9QaecBHW4gPGDxXcmVt7Ja", decimal.NewFromFloat(0.94999)},
	}

	client := rpc.New(rpc.TestNet_RPC)

	for _, i := range tests {
		balance, err := GetBalanceSol(context.TODO(), client, i.address)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s: %v: %v\n", i.address, i.balance, balance)

	}

}

func TestSendTx(t *testing.T) {
	client := rpc.New(rpc.TestNet_RPC)
	ws, err := ws.Connect(context.Background(), rpc.TestNet_WS)
	if err != nil {
		t.Fatal(err)
	}

	priv, err := StringToPriv("4dU151QugGbkcq6x5ENShjWV2doKf9LBBKF1S7nxi47dr1qCKRtTKxmy41ScCUFhTPBXEC2yWCZQkVBW1xLjKAUY")
	if err != nil {
		t.Fatal(err)
	}

	println(priv.PublicKey().String())

	tx, _, err := CreateTx(client, 20_000_000, priv, solana.MustPublicKeyFromBase58("9rVz3dfegDTh5Wf1tD7y4k9QaecBHW4gPGDxXcmVt7Ja"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Created tx")

	tx = tx
	ws = ws
	// sig, err := SendTx(client, ws, tx)
	// fmt.Println(sig, err)

	// t.Log("Sent:", sig.String())

}

func TestGetTxStatus(t *testing.T) {
	client := rpc.New(rpc.TestNet_RPC)

	tests := []struct {
		hash   string
		amount decimal.Decimal
	}{
		{"2Uh6RpU1aVDi3HKUVETZjQh3ReSX8LEwcmgrs2SxZQwGmi3XncUrSmjXpmnXb7AQhEuXRn3n7QUDSAXPoWm8DEPp", decimal.NewFromInt(20)},
	}

	tests = tests

	amount, success, err := GetTxByHash(client, "57XhG2CwWEDccZYKJTv1zd3DhbDsDe7nX9inuqo6391iVBw12JRGt5roDo96j2XouTN6spnG1eevrsSQ12YmyAxq")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(amount, success)

}

func TestCreateWallet2(t *testing.T) {
	addr, priv, _ := NewWallet()

	fmt.Println("Address: ", addr)
	fmt.Println("PrivateKey: ", priv)

}
