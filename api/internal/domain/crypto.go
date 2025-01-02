package domain

type Crypto uint8

const (
	CRYPTO_NONE Crypto = iota // only for init
	CRYPTO_SOL
	CRYPTO_TON
	CRYPTO_ETH
)

var Cryptos = [...]string{"none", "sol", "ton", "eth"}

func (c Crypto) ToString() string {
	return Cryptos[c]
}

func (c Crypto) IsNone() bool {
	return c == 0
}

func StrToCrypto(s string) Crypto {
	for i, currencyName := range Cryptos {
		if s == currencyName {
			return Crypto(i)
		}
	}
	return CRYPTO_NONE
}
