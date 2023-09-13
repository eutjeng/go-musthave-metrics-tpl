package utils

import (
	"strconv"
	"strings"
	"time"
)

func SplitPath(path string) []string {
	return strings.Split(strings.TrimPrefix(path, "/"), "/")
}

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ConvertToSec(s string) (time.Duration, error) {
	sec, err := ParseInt(s)

	if err != nil {
		return 0, err
	}

	return time.Duration(sec) * time.Second, nil
}

func EnsureHTTPScheme(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}
