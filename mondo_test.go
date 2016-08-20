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
	baseEndpoint = servUrl.String()

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

	client, err = Authenticate("", "", "", "")
	assert.Error(t, err)
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
			            "settled": "2015-08-23T12:20:18Z",
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
			            "settled": "2015-08-23T12:20:18Z",
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

func TestAuthentication(t *testing.T) {
	setup()
	defer teardown()

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)
	assert.True(t, client.Authenticated())

	client.expiryTime = time.Now()
	assert.False(t, client.Authenticated())
}

func TestTransactionByID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/transactions/transaction1",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"transaction": {
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
			            "settled": "2015-08-23T12:20:18Z",
			            "category": "eating_out"
			        }}
`)
		},
	)

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)

	transaction, err := client.TransactionByID("account1", "transaction1")
	assert.NoError(t, err)
	assert.Equal(t, -510, transaction.Amount)
}

func TestAccounts(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{
									    "accounts": [
									        {
									            "id": "acc_00009237aqC8c5umZmrRdh",
									            "description": "Peter Pan's Account",
									            "created": "2015-11-13T12:17:42Z"
									        }
									    ]
									}`)
		},
	)

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)

	accounts, err := client.Accounts()
	assert.NoError(t, err)

	assert.Equal(t, 1, len(accounts))

	account := accounts[0]
	assert.Equal(t, "Peter Pan's Account", account.Description)
	assert.Equal(t, "2015-11-13T12:17:42Z", account.Created.Format(time.RFC3339))
}

func TestCreateItem(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/feed",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Hello!", r.FormValue("params[title]"))
			assert.Equal(t, "account1", r.FormValue("account_id"))
			assert.Equal(t, "http://www.gophers.com/gopher1.png", r.FormValue("params[image_url]"))
			assert.Equal(t, "A body goes here", r.FormValue("params[body]"))
			fmt.Fprint(w, `{}`)
		},
	)

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)

	err = client.CreateFeedItem("account1", &FeedItem{
		Title:    "Hello!",
		ImageURL: "http://www.gophers.com/gopher1.png",
		Body:     "A body goes here",
	})
	assert.NoError(t, err)

	err = client.CreateFeedItem("", nil)
	assert.Error(t, err)
}

func TestRegisterWebhook(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/webhooks",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "http://www.google.com", r.FormValue("url"))
			fmt.Fprint(w, `{
									    "webhook": {
									        "account_id": "account_id",
									        "id": "webhook_id",
									        "url": "http://www.google.com"
									    }
									}`)
		},
	)

	client, err := Authenticate("some", "valid", "credentials", "here")
	assert.NoError(t, err)

	webhook, err := client.RegisterWebhook("account1", "http://www.google.com")
	assert.NoError(t, err)

	assert.Equal(t, "http://www.google.com", webhook.Url)
}
