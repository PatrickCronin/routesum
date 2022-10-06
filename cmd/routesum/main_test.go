package main

import (
	"io"
	"regexp"
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
			expected:     regexp.MustCompile(`Num internal nodes`),
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

			err := summarize(in, &out, memStatsOut, nil)
			require.NoError(t, err, "summarize does not throw an error")

			assert.Equal(t, "192.0.2.0/31\n", out.String(), "read expected output")
			assert.Regexp(t, test.expected, memStatsBuilder.String(), "read expected memory stats")
		})
	}
}
