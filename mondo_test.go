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

	// Bake an auth request into the server to return a client
	mux.HandleFunc("/oauth2/token",
		func(w http.ResponseWriter, r *http.Request) {
			if val := r.FormValue("client_secret"); val != "valid" {
				w.WriteHeader(401)
				return
			}
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
}

func teardown() {
	mux = nil
	server = nil
}

func TestOAuth(t *testing.T) {
	setup()
	defer teardown()

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)

	assert.NotNil(t, client)
	assert.False(t, client.ExpiresAt().Before(time.Now()))
	assert.True(t, client.authenticated)

	client, err = Authenticate("some", "notvalid", "credentials", "here")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, ErrUnauthenticatedRequest, err)
}

func TestTransactions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/transactions",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"transactions": [{
			            "account_balance": 13013,
			            "amount": -510,
			            "created": "2015-08-22T12:20:18Z",
			            "currency": "GBP",
			            "description": "THE DE BEAUVOIR DELI C LONDON        GBR",
			            "id": "tx_00008zIcpb1TB4yeIFXMzx",
									"merchant": {
											"address": {
												"address": "98 Southgate Road",
												"city": "London",
												"country": "GB",
												"latitude": 51.54151,
												"longitude": -0.08482400000002599,
												"postcode": "N1 3JD",
												"region": "Greater London"
											},
											"created": "2015-08-22T12:20:18Z",
											"group_id": "grp_00008zIcpbBOaAr7TTP3sv",
											"id": "merch_00008zIcpbAKe8shBxXUtl",
											"logo": "https://pbs.twimg.com/profile_images/527043602623389696/68_SgUWJ.jpeg",
											"emoji": "🍞",
											"name": "The De Beauvoir Deli Co.",
											"category": "eating_out"
										},
			            "metadata": {},
			            "notes": "Salmon sandwich 🍞",
			            "is_load": false,
			            "settled": true,
			            "category": "eating_out"
			        },
			        {
			            "account_balance": 12334,
			            "amount": -679,
			            "created": "2015-08-23T16:15:03Z",
			            "currency": "GBP",
			            "description": "VUE BSL LTD            ISLINGTON     GBR",
			            "id": "tx_00008zL2INM3xZ41THuRF3",
									"merchant": {
											"address": {
												"address": "98 Southgate Road",
												"city": "London",
												"country": "GB",
												"latitude": 51.54151,
												"longitude": -0.08482400000002599,
												"postcode": "N1 3JD",
												"region": "Greater London"
											},
											"created": "2015-08-22T12:20:18Z",
											"group_id": "grp_00008zIcpbBOaAr7TTP3sv",
											"id": "merch_00008zIcpbAKe8shBxXUtl",
											"logo": "https://pbs.twimg.com/profile_images/527043602623389696/68_SgUWJ.jpeg",
											"emoji": "🍞",
											"name": "The De Beauvoir Deli Co.",
											"category": "eating_out"
										},
			            "metadata": {},
			            "notes": "",
			            "is_load": false,
			            "settled": true,
			            "category": "eating_out"
			        }]}
`)
		},
	)

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)
	transactions, err := client.Transactions("an account", "", "", 100)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(transactions))

	assert.Equal(t, transactions[0].Currency, "GBP")
	assert.Equal(t, transactions[0].Merchant.Emoji, "🍞")
}
