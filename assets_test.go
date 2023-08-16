package spa_test

import (
	"bytes"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dio/spa"
	"github.com/stretchr/testify/require"
)

// Explicitly adding _* is required: https://github.com/golang/go/issues/42328#issuecomment-736971637.
//
//go:embed testdata/app testdata/app/statics/js/_*
var testdataFs embed.FS

func TestServeHTTP(t *testing.T) {
	assets, err := spa.NewAssets(testdataFs, "testdata/app")
	require.NoError(t, err)

	tests := []struct {
		path     string
		validate func(body *bytes.Buffer)
	}{
		{
			path: "/",
			validate: func(body *bytes.Buffer) {
				requireContainsBodyString(t, body, "<head>")
			},
		},
		{
			path: "/manifest.json",
			validate: func(body *bytes.Buffer) {
				requireContainsBodyString(t, body, "{}")
			},
		},
		{
			path: "/statics/ok.json",
			validate: func(body *bytes.Buffer) {
				requireContainsBodyString(t, body, "{}")
			},
		},
		{
			path: "/statics/js/_baseUniq.a46ea275.js",
			validate: func(body *bytes.Buffer) {
				requireContainsBodyString(t, body, "console.log('base');")
			},
		},
		{
			path: "/statics/js/baseUniq.a46ea275.js",
			validate: func(body *bytes.Buffer) {
				requireContainsBodyString(t, body, "console.log('base');")
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.path, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, test.path, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(assets.ServeHTTP)
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusOK)
			if test.validate != nil {
				test.validate(rr.Body)
			}
		})
	}

}

func requireContainsBodyString(t *testing.T, body *bytes.Buffer, expected string) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Contains(t, string(b), expected)
}
