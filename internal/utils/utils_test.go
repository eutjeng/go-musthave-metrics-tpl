package utils

import (
	"strconv"
	"testing"

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
