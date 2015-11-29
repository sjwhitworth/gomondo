package mondo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	servUrl, _ := url.Parse(server.URL)
	BaseMondoURL = servUrl.String()
}

func teardown() {
	mux = nil
	server = nil
}

func TestOAuth(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/oauth2/token",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "some", r.FormValue("client_id"))
			assert.Equal(t, "valid", r.FormValue("client_secret"))
			assert.Equal(t, "credentials", r.FormValue("username"))
			assert.Equal(t, "here", r.FormValue("password"))
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			assert.Equal(t, "POST", r.Method)
			fmt.Fprint(w, `{
                      "access_token": "access_token",
                      "client_id": "client_id",
                      "expires_in": 21600,
                      "refresh_token": "refresh_token",
                      "token_type": "Bearer",
                      "user_id": "user_id"
                      }`)
		},
	)

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)

	assert.NotNil(t, client)
	assert.False(t, client.ExpiresAt().Before(time.Now()))
	assert.True(t, client.authenticated)
}
