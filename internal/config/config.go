package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	FlagRunAddr           string
	FlagRunReportInterval int64
	FlagRunPollInterval   int64

	ReportInterval time.Duration
	PollInterval   time.Duration

	PollCount   int64
	RandomValue float64
)

func customUsage(flagSet *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	flagSet.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Error: Unknown or missing flag(s) provided\n")
	os.Exit(2)
}

func registerFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flagSet.Int64Var(&FlagRunReportInterval, "r", 10, "frequency of sending metrics to the server (seconds)")
	flagSet.Int64Var(&FlagRunPollInterval, "p", 2, "frequency of metrics polling from the runtime package (seconds)")
}

func parseFlags(flagSet *flag.FlagSet) error {
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return err
	}
	return nil
}

func applyParsedValues() {
	ReportInterval = time.Duration(FlagRunReportInterval) * time.Second
	PollInterval = time.Duration(FlagRunPollInterval) * time.Second
}

func ParseFlags() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.Usage = func() { customUsage(flagSet) }

	registerFlags(flagSet)

	if err := parseFlags(flagSet); err != nil {
		return
	}

	applyParsedValues()
}
