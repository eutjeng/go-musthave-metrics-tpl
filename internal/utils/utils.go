package utils

import (
	"strconv"
	"strings"
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

func EnsureHTTPScheme(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}
