package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecureHeaders(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, r)

	rs := rr.Result()

	frameOptions := rs.Header.Get(headerFrameOptions)
	require.Equal(t, headerFrameOptionsValue, frameOptions)

	xssProtection := rs.Header.Get(headerXSSProtection)
	require.Equal(t, headerXSSProtectionValue, xssProtection)

	require.Equal(t, http.StatusOK, rs.StatusCode)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "OK", string(body))
}
