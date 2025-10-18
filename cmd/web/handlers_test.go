package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	rs := ts.get(t, "/ping")

	require.Equal(t, http.StatusOK, rs.statusCode)
	require.Equal(t, "OK", rs.body)
}

func TestShowSnippet(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, "Content"},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, ""},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, ""},
		{"Decimal ID", "/snippet/1.23", http.StatusNotFound, ""},
		{"String ID", "/snippet/foo", http.StatusNotFound, ""},
		{"Empty ID", "/snippet/", http.StatusNotFound, ""},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ts.get(t, tt.urlPath)
			require.Equal(t, tt.wantCode, response.statusCode)
			require.True(t, strings.Contains(response.body, tt.wantBody))
		})
	}
}
