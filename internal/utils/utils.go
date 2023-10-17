package utils

import (
	"fmt"
	"reflect"
	"sort"
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

// getSortedKeys takes a map and returns its keys sorted as a slice of strings
func getSortedKeys(m interface{}) []string {
	var keys []string
	v := reflect.ValueOf(m)

	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			keys = append(keys, key.String())
		}

		sort.Strings(keys)
	}

	return keys
}

// formatMapSortedKeys formats the key-value pairs of a map to a string,
// with keys sorted
func FormatMapSortedKeys(m interface{}) string {
	var result strings.Builder
	keys := getSortedKeys(m)

	for _, key := range keys {
		switch mapType := m.(type) {
		case map[string]int64:
			result.WriteString(fmt.Sprintf("%s: %d\n", key, mapType[key]))
		case map[string]float64:
			result.WriteString(fmt.Sprintf("%s: %f\n", key, mapType[key]))
		}
	}
	return result.String()
}
