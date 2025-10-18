package main

import (
	"html"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/stretchr/testify/require"
	"github.com/yoqub-davlatov/snippetbox/pkg/models/mock"
)

var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

type testServer struct {
	*httptest.Server
}

type testServerResponse struct {
	statusCode int
	headers    http.Header
	body       string
}

func newTestApplication(t *testing.T) *application {

	templateCache, err := newTemplateCache("./../../ui/html/")
	if err != nil {
		t.Fatal(err)
	}

	session := sessions.New([]byte("3dSm5MnygFHh7XidAtbskXrjbwfoJcbJ"))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	return &application{
		// errorLog:      log.New(io.Discard, "", 0),
		// infoLog:       log.New(io.Discard, "", 0),
		infoLog:       log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog:      log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		templateCache: templateCache,
		session:       session,
		users:         &mock.UserModel{},
		snippets:      &mock.SnippetModel{},
	}
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	ts.Client().Jar = jar
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) testServerResponse {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	require.NoError(t, err)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	require.NoError(t, err)

	return testServerResponse{
		statusCode: rs.StatusCode,
		headers:    rs.Header,
		body:       string(body),
	}
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) testServerResponse {
	req, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, strings.NewReader(form.Encode()))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	rs, err := ts.Client().Do(req)
	require.NoError(t, err)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	require.NoError(t, err)

	return testServerResponse{
		statusCode: rs.StatusCode,
		headers:    rs.Header,
		body:       string(body),
	}
}

func extractCSRFToken(t *testing.T, body []byte) string {
	matches := csrfTokenRX.FindSubmatch(body)
	require.GreaterOrEqual(t, len(matches), 2)

	return html.UnescapeString(string(matches[1]))
}
