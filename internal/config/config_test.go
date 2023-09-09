package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		args                   []string
		expectedRunAddr        string
		expectedReportInterval int64
		expectedPollInterval   int64
	}{
		{[]string{"-a", ":8081", "-r", "20", "-p", "3"}, ":8081", 20, 3},
		{[]string{"-a", ":8082"}, ":8082", 10, 2},
		{[]string{"-r", "15"}, ":8080", 15, 2},
		{[]string{"-p", "4"}, ":8080", 10, 4},
		{[]string{}, ":8080", 10, 2},
		{[]string{"-a", ":8083", "-r", "25"}, ":8083", 25, 2},
	}

	for _, test := range tests {
		os.Args = append([]string{"cmd"}, test.args...)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		ParseFlags()

		if FlagRunAddr != test.expectedRunAddr {
			t.Errorf("expected %s, got %s", test.expectedRunAddr, FlagRunAddr)
		}

		if FlagRunReportInterval != test.expectedReportInterval {
			t.Errorf("expected %d, got %d", test.expectedReportInterval, FlagRunReportInterval)
		}

		if FlagRunPollInterval != test.expectedPollInterval {
			t.Errorf("expected %d, got %d", test.expectedPollInterval, FlagRunPollInterval)
		}

		if ReportInterval != time.Duration(test.expectedReportInterval)*time.Second {
			t.Errorf("expected %s, got %s", time.Duration(test.expectedReportInterval)*time.Second, ReportInterval)
		}

		if PollInterval != time.Duration(test.expectedPollInterval)*time.Second {
			t.Errorf("expected %s, got %s", time.Duration(test.expectedPollInterval)*time.Second, PollInterval)
		}
	}
}
