package currencies

import (
	"github.com/shopspring/decimal"
)

func CalculateCommission(amount decimal.Decimal, __commission float64) (finalAmount decimal.Decimal, commissionAmount decimal.Decimal) {
	if __commission == 0 { // zero commission
		return amount, decimal.NewFromInt(0)
	}

	// amount := decimal.NewFromFloat(0.10000)

	commissionAmount = amount.Mul(decimal.NewFromFloat(__commission)).Div(decimal.NewFromInt(100))
	finalAmount = amount.Sub(commissionAmount)

	return finalAmount, commissionAmount
}
