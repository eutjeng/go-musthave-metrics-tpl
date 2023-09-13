package utils

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSplitPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected []string
	}{
		{"/a/b/c", []string{"a", "b", "c"}},
		{"/a//c", []string{"a", "", "c"}},

		{"a/b/c", []string{"a", "b", "c"}},

		{"/", []string{""}},

		{"", []string{""}},
	}

	for _, tc := range testCases {
		result := SplitPath(tc.path)
		assert.Equal(t, tc.expected, result)
	}
}
func TestParseFloat(t *testing.T) {
	testCases := []struct {
		s        string
		expected float64
		err      error
	}{
		{"3.14", 3.14, nil},
		{"0", 0, nil},
		{"-3.14", -3.14, nil},
		{"abc", 0, &strconv.NumError{Func: "ParseFloat", Num: "abc", Err: strconv.ErrSyntax}},
	}

	for _, tc := range testCases {
		result, err := ParseFloat(tc.s)
		assert.Equal(t, tc.expected, result)
		assert.Equal(t, tc.err, err)
	}
}

func TestParseInt(t *testing.T) {
	testCases := []struct {
		s        string
		expected int64
		err      error
	}{
		{"3", 3, nil},
		{"0", 0, nil},
		{"-3", -3, nil},
		{"abc", 0, &strconv.NumError{Func: "ParseInt", Num: "abc", Err: strconv.ErrSyntax}},
	}

	for _, tc := range testCases {
		result, err := ParseInt(tc.s)
		assert.Equal(t, tc.expected, result)
		assert.Equal(t, tc.err, err)
	}
}

func TestConvertToSec(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"10", 10 * time.Second, false},
		{"0", 0, false},
		{"-5", -5 * time.Second, false},
		{"abc", 0, true}, // ожидается ошибка
		{"", 0, true},    // ожидается ошибка
	}

	for _, test := range tests {
		result, err := ConvertToSec(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("ConvertToSec(%s) expected error, got nil", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("ConvertToSec(%s) got error: %v", test.input, err)
			continue
		}

		if result != test.expected {
			t.Errorf("ConvertToSec(%s) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestEnsureHTTPScheme(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://localhost", "http://localhost"},
		{"https://localhost", "https://localhost"},
		{"localhost", "http://localhost"},
		{"ftp://localhost", "http://ftp://localhost"},
	}

	for _, test := range tests {
		result := EnsureHTTPScheme(test.input)
		if result != test.expected {
			t.Errorf("EnsureHTTPScheme(%s) = %s, want %s", test.input, result, test.expected)
		}
	}
}
