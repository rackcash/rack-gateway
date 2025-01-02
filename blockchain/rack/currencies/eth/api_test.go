package eth

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
)

func TestWeiToEther(t *testing.T) {

	tests := []struct {
		wei   *big.Int
		ether decimal.Decimal
	}{
		{big.NewInt(20000000000000), decimal.NewFromFloat(0.00002)},             // ~ 0.07 usd
		{big.NewInt(572000000000000), decimal.NewFromFloat(0.000572)},           // ~ 2 usd
		{big.NewInt(1096032890738000), decimal.NewFromFloat(0.001096032890738)}, // ~ 4 usd
		{big.NewInt(1303000000000000), decimal.NewFromFloat(0.001303)},          // ~ 5 usd
		{big.NewInt(50000000000000000), decimal.NewFromFloat(0.05)},             // ~ 170 usd
		{big.NewInt(5000000000000000000), decimal.NewFromFloat(5)},              // ~ 17000 usd
	}

	for _, i := range tests {
		ether := WeiToEther(i.wei)
		if !ether.Equal(i.ether) {
			t.Fatalf("WeiToEther(%s) = %s, want %s", i.wei, ether, i.ether)
		}
	}

}

func TestCreateTx(t *testing.T) {

	client := Connect("https://holesky.infura.io/v3/" /* infura api key */)

	tx, err := CreateTx(client, "0xB6feDEF89B7311AD816d7D8Ecea8237C9591b5D0", "0xB6feDEF89B7311AD816d7D8Ecea8237C9591b5D0", decimal.NewFromInt(100).BigInt(), 10)
	if err != nil {
		t.Fatal(err)
	}
	tx = tx

}

func TestGetBalanceEther(t *testing.T) {
	client := Connect("https://holesky.infura.io/v3/")

	e, err := GetBalanceEther(client, "0x6232f7d132e2a0Db8AA6cD7bbE587A0FEEa620a5")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(e)
}
