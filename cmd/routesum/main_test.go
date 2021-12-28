package main

import (
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarize(t *testing.T) {
	tests := []struct {
		name         string
		showMemStats bool
		expected     *regexp.Regexp
	}{
		{
			name:         "without memory statistics",
			showMemStats: false,
			expected:     regexp.MustCompile(`^$`),
		},
		{
			name:         "with memory statistics",
			showMemStats: true,
			expected: regexp.MustCompile(
				`Before work(?:.|\n)+After building the summary(?:.|\n)+After writing the summary`,
			),
		},
	}

	inStr := "\n192.0.2.0\n192.0.2.1\n"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := strings.NewReader(inStr)
			var out strings.Builder

			var memStatsBuilder strings.Builder
			var memStatsOut io.Writer

			if test.showMemStats {
				memStatsOut = &memStatsBuilder
			}

			err := summarize(in, &out, memStatsOut)
			require.NoError(t, err, "summarize does not throw an error")

			assert.Equal(t, "192.0.2.0/31\n", out.String(), "read expected output")
			assert.Regexp(t, test.expected, memStatsBuilder.String(), "read expected memory stats")
		})
	}
}

func TestFormatByteCount(t *testing.T) {
	tests := []struct {
		value    int64
		expected string
	}{
		{
			value:    82,
			expected: "82 B",
		},
		{
			value:    1024,
			expected: "1.0 KB",
		},
		{
			value:    2000000,
			expected: "1.9 MB",
		},
	}

	for _, test := range tests {
		t.Run(strconv.FormatInt(test.value, 10), func(t *testing.T) {
			assert.Equal(t, test.expected, formatByteCount(test.value))
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		value    uint64
		expected string
	}{
		{
			value:    82,
			expected: "82",
		},
		{
			value:    1024,
			expected: "1,024",
		},
		{
			value:    1234567890,
			expected: "1,234,567,890",
		},
	}

	for _, test := range tests {
		t.Run(strconv.FormatUint(test.value, 10), func(t *testing.T) {
			assert.Equal(t, test.expected, formatNumber(test.value))
		})
	}
}
