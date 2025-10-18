package main

import (
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/stretchr/testify/require"
	"github.com/yoqub-davlatov/snippetbox/pkg/models/mock"
)

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

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		templateCache: templateCache,
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
