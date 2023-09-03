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
