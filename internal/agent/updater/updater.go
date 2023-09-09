package updater

import (
	"math/rand"
)

func UpdateMetrics(pollCount *int64, randomValue *float64) {
	*pollCount++
	*randomValue = rand.Float64()
}
