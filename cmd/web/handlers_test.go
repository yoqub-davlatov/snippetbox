package main

import (
	"net/http"
	"net/url"
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

func TestSignupUser(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	signupForm := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, []byte(signupForm.body))
	t.Log(csrfToken)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     string
	}{
		{"Valid submission", "Bob", "bob@example.com", "validPa$$word", csrfToken, http.StatusSeeOther, ""},
		{"Empty name", "", "bob@example.com", "validPa$$word", csrfToken, http.StatusOK, "This field cannot be blank"},
		{"Empty email", "Bob", "", "validPa$$word", csrfToken, http.StatusOK, "This field cannot be blank"},
		{"Empty password", "Bob", "bob@example.com", "", csrfToken, http.StatusOK, "This field cannot be blank"},
		{"Invalid email (missing @)", "Bob", "bobexample.com", "validPa$$word", csrfToken, http.StatusOK, "This field is invalid"},
		{"Invalid email (missing local part)", "Bob", "@example.com", "validPa$$word", csrfToken, http.StatusOK, "This field is invalid"},
		{"Short password", "Bob", "bob@example.com", "pa$$word", csrfToken, http.StatusOK, "This field is too short (minimum is 10 characters)"},
		{"Duplicate email", "Bob", "duplicate@example.com", "validPa$$word", csrfToken, http.StatusOK, "Address is already in use"},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)
			response := ts.postForm(t, "/user/signup", form)
			require.Equal(t, response.statusCode, tt.wantCode)
			require.True(t, strings.Contains(response.body, tt.wantBody))
		})
	}
}
