package config

import (
	"os"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		expectedAddr           string
		envVars                map[string]string
		args                   []string
		expectedReportInterval time.Duration
		expectedPollInterval   time.Duration
	}{
		{
			envVars: map[string]string{
				"ADDRESS":         ":8081",
				"REPORT_INTERVAL": "20",
				"POLL_INTERVAL":   "3",
			},
			args:                   []string{"-a", ":8081", "-r", "20", "-p", "3"},
			expectedAddr:           ":8081",
			expectedReportInterval: 20 * time.Second,
			expectedPollInterval:   3 * time.Second,
		},
		{
			envVars:                nil,
			args:                   []string{"-a", ":8082"},
			expectedAddr:           ":8082",
			expectedReportInterval: 10 * time.Second,
			expectedPollInterval:   2 * time.Second,
		},
		{
			envVars: map[string]string{
				"ADDRESS": ":8083",
			},
			args:                   nil,
			expectedAddr:           ":8083",
			expectedReportInterval: 10 * time.Second,
			expectedPollInterval:   2 * time.Second,
		},
		{
			envVars: map[string]string{
				"REPORT_INTERVAL": "25",
			},
			args:                   []string{"-p", "5"},
			expectedAddr:           ":8080",
			expectedReportInterval: 25 * time.Second,
			expectedPollInterval:   5 * time.Second,
		},
		{
			envVars:                nil,
			args:                   nil,
			expectedAddr:           ":8080",
			expectedReportInterval: 10 * time.Second,
			expectedPollInterval:   2 * time.Second,
		},
	}

	for _, test := range tests {
		for k, v := range test.envVars {
			os.Setenv(k, v)
		}

		os.Args = append([]string{"cmd"}, test.args...)
		cfg, err := ParseConfig()
		if err != nil {
			t.Errorf("ParseConfig failed: %s", err)
			continue
		}

		if cfg.Addr != test.expectedAddr {
			t.Errorf("expected %s, got %s", test.expectedAddr, cfg.Addr)
		}
		if cfg.ReportInterval != test.expectedReportInterval {
			t.Errorf("expected %s, got %s", test.expectedReportInterval, cfg.ReportInterval)
		}
		if cfg.PollInterval != test.expectedPollInterval {
			t.Errorf("expected %s, got %s", test.expectedPollInterval, cfg.PollInterval)
		}

		for k := range test.envVars {
			os.Unsetenv(k)
		}
	}
}
