package reporter

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/utils"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/hash"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func sendRequestWithHashing(cfg *config.Config, sugar *zap.SugaredLogger, client *resty.Client, url string, compressedBody []byte, hash string) error {
	sugar.Infof("request hash: %s", hash)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("HashSHA256", hash).
		SetBody(compressedBody).
		Post(url)

	if err != nil {
		sugar.Errorw("error sending request for metric", "error", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		sugar.Errorw("received non-OK response for metric", "status", resp.Status())
		return fmt.Errorf("received non-OK response: %s", resp.Status())
	}

	return nil
}

func ReportMetrics(cfg *config.Config, sugar *zap.SugaredLogger, url string, client *resty.Client, res []models.Metrics) error {
	sugar.Infof("Sending metrics to %s: %v", url, res)
	jsonData, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("json marshaling failed: %w", err)
	}

	hash := hash.ComputeHash(jsonData, cfg.Key)

	compressedData, err := utils.CompressData(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress json data: %w", err)
	}

	err = sendRequestWithHashing(cfg, sugar, client, url, compressedData, hash)
	if err != nil {
		return fmt.Errorf("failed to send request with hashing: %w", err)
	}

	return nil
}
