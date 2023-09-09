package updater

import (
	"math/rand"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
)

func UpdateMetrics() {
	config.PollCount++
	config.RandomValue = rand.Float64()
}
