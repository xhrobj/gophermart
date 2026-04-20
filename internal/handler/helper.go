package handler

import (
	"errors"
	"math"
)

func hundredthsToAmount(v int64) float64 {
	return float64(v) / 100
}

func amountToHundredths(amount float64) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, errors.New("invalid amount")
	}

	return int64(math.Round(amount * 100)), nil
}
