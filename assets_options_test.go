package spa_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefix(t *testing.T) {
	tests := []struct {
		input    string
		replacer string
		output   string
	}{
		{"/%DEPLOYMENT_PATH%/favicon.ico", "", "/favicon.ico"},
		{"/%DEPLOYMENT_PATH%/static/favicon.ico", "", "/static/favicon.ico"},
		{"/%DEPLOYMENT_PATH%/favicon.ico", "ok", "/ok/favicon.ico"},
		{"/%DEPLOYMENT_PATH%/favicon.ico", "ok/cool", "/ok/cool/favicon.ico"},
	}

	for _, testCase := range tests {
		testCase := testCase
		pattern := "%DEPLOYMENT_PATH%"
		if testCase.replacer == "" {
			pattern = pattern + "/"
		}
		assert.Equal(t, testCase.output, strings.Replace(testCase.input, pattern, testCase.replacer, 1))
	}
}
