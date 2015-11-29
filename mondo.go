// Package go-mondo provides a Go interface for interacting with the Mondo API.
package mondo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// The root URL we will base all queries off of. Currently only production is supported.
	BaseMondoURL = "https://production-api.gmon.io"

	// OAuth grant type.
	GrantTypePassword = "password"

	// 401 response code
	ErrUnauthenticatedRequest = fmt.Errorf("your request was not sent with a valid token")
)

type MondoClient struct {
	accessToken   string
	authenticated bool
	expiryTime    time.Time
}

// Function Authenticate authenticates the user using the oath flow, returning an authenticated MondoClient
func Authenticate(clientId, clientSecret, username, password string) (*MondoClient, error) {
	if clientId == "" || clientSecret == "" || username == "" || password == "" {
		return nil, fmt.Errorf("zero value passed to Authenticate")
	}

	values := url.Values{}
	values.Set("grant_type", GrantTypePassword)
	values.Set("client_id", clientId)
	values.Set("client_secret", clientSecret)
	values.Set("username", username)
	values.Set("password", password)

	resp, err := http.PostForm(buildUrl("oauth2/token"), values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tresp := tokenResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &tresp); err != nil {
		return nil, err
	}

	if tresp.Error != "" {
		return nil, fmt.Errorf(tresp.Error)
	}

	if tresp.ExpiresIn == 0 || tresp.TokenType == "" || tresp.AccessToken == "" {
		return nil, fmt.Errorf("failed to scan response correctly")
	}

	return &MondoClient{
		authenticated: true,
		accessToken:   tresp.AccessToken,
		expiryTime:    time.Now().Add(time.Duration(tresp.ExpiresIn) * time.Second),
	}, nil
}

// ExpiresAt returns the time that the current oauth token expires and will have to be refreshed.
func (m *MondoClient) ExpiresAt() time.Time {
	return m.expiryTime
}

// callWithAuth makes authenticated calls to the Mondo API.
func (m *MondoClient) callWithAuth(methodType, URL string, params map[string]string) (*http.Response, error) {
	var resp *http.Response
	var err error

	// TODO: This is so hacky, clean up
	switch methodType {
	case "GET":
		req, err := http.NewRequest(methodType, buildUrl(URL), nil)
		if err != nil {
			return nil, err
		}

		// If we have any parameters, add them here.
		if len(params) > 0 {
			query := req.URL.Query()
			for k, v := range params {
				query.Add(k, v)
			}
			req.URL.RawQuery = query.Encode()
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			return nil, ErrUnauthenticatedRequest
		}

	case "POST":
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}

		req, err := http.NewRequest(methodType, buildUrl(URL), strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			return nil, ErrUnauthenticatedRequest
		}
	}

	return resp, err
}

// Transactions returns a slice of Transactions, with the merchant expanded within the Transaction. This endpoint supports pagination. To paginate, provide the last Transacation.ID to the since parameter of the function, if the length of the results that are returned is equal to your limit.
func (m *MondoClient) Transactions(accountId, since, before string, limit int) ([]Transaction, error) {
	type transactionsResponse struct {
		Transactions []Transaction `json:"transactions"`
	}

	params := map[string]string{
		"account_id": accountId,
		"expand[]":   "merchant",
		"limit":      fmt.Sprintf("%v", limit),
		"since":      since,
		"before":     before,
	}

	resp, err := m.callWithAuth("GET", fmt.Sprintf("transactions"), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tresp := transactionsResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &tresp); err != nil {
		return nil, err
	}

	return tresp.Transactions, nil
}

// TransactionByID obtains a Mondo Transaction by a specific transaction ID.
func (m *MondoClient) TransactionByID(accountId, transactionId string) (*Transaction, error) {
	type transactionByIDResponse struct {
		Transaction Transaction `json:"transaction"`
	}

	params := map[string]string{
		"account_id": accountId,
		"expand[]":   "merchant",
	}

	resp, err := m.callWithAuth("GET", fmt.Sprintf("transactions/%s", transactionId), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tresp := transactionByIDResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &tresp); err != nil {
		return nil, err
	}

	return &tresp.Transaction, nil
}

func (m *MondoClient) Accounts() ([]Account, error) {
	type accountsResponse struct {
		Accounts []Account `json:"accounts"`
	}

	resp, err := m.callWithAuth("GET", "accounts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	acresp := accountsResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &acresp); err != nil {
		return nil, err
	}

	return acresp.Accounts, nil
}

// CreateFeedItem creates a feed item in the user's application.
// TODO: There is no way to delete a feed item currently, so use with caution.
func (m *MondoClient) CreateFeedItem(accountId, title, imageURL, bgColor, bodyColor, titleColor, body string) error {
	type feedItemResponse struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if imageURL == "" {
		return fmt.Errorf("imageURL cannot be empty")
	}

	if accountId == "" {
		return fmt.Errorf("accountId cannot be empty")
	}

	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if body == "" {
		return fmt.Errorf("body cannot be empty")
	}

	if bgColor == "" {
		bgColor = "#FCF1EE"
	}

	if bodyColor == "" {
		bodyColor = "#FCF1EE"
	}

	if titleColor == "" {
		titleColor = "#333"
	}

	params := map[string]string{
		"account_id":               accountId,
		"type":                     "basic",
		"params[title]":            title,
		"params[image_url]":        imageURL,
		"params[background_color]": bgColor,
		"params[body_color]":       bodyColor,
		"params[title_color]":      titleColor,
		"params[body]":             body,
	}

	resp, err := m.callWithAuth("POST", fmt.Sprintf("feed"), params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var fresp feedItemResponse
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &fresp); err != nil {
		return err
	}

	// Generate a nicely formatted error code back to the caller
	if fresp.Code != "" {
		return fmt.Errorf("%v: %v", fresp.Code, fresp.Message)
	}

	return nil
}

func (m *MondoClient) RegisterWebhook(URL string) error {
	return nil
}

func buildUrl(path string) string {
	return fmt.Sprintf("%v/%v", BaseMondoURL, path)
}
