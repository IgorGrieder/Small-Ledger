package application

// Fee of 1%
const GATEWAY_FEE int64 = 1

func CalculateFee(amount int64) int64 {
	return (amount * GATEWAY_FEE) / 10000
}
