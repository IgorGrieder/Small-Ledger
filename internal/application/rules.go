package application

import "github.com/google/uuid"

// Fee of 1%
const GATEWAY_FEE int64 = 1

func CalculateFee(amount int64) int64 {
	return (amount * GATEWAY_FEE) / 10000
}

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyBRL Currency = "BRL"
)

// Fixed Systems Pools
var (
	SystemPoolUSD = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	SystemPoolBRL = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)
