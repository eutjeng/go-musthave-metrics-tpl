package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"math/rand"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
)

const URLTemplate = "%s/updates"

func UpdateMetrics(pollCount *int64, randomValue *float64) {
	*pollCount++
	*randomValue = rand.Float64()
}

func GenerateMetricURL(addr string) string {
	return fmt.Sprintf(URLTemplate, utils.EnsureHTTPScheme(addr))
}

func CompressData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
