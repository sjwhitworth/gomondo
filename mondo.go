// Package go-mondo provides a Go wrapper for the Mondo API.
package mondo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// The root URL
	baseEndpoint = "https://production-api.gmon.io"

	grantTypePassword = "password"

	// 401 response code
	ErrUnauthenticatedRequest = errors.New("mondo: your request was not sent with a valid token")

	// No transaction found
	ErrNoTransactionFound = errors.New("mondo: no transaction found with ID")
)

type Client struct {
	accessToken   string
	authenticated bool
	expiryTime    time.Time
}

// Function Authenticate authenticates the user using the oath flow, returning an authenticated Client
func Authenticate(clientId, clientSecret, username, password string) (*Client, error) {
	if clientId == "" || clientSecret == "" || username == "" || password == "" {
		return nil, fmt.Errorf("zero value passed to Authenticate")
	}

	values := url.Values{}
	values.Set("grant_type", grantTypePassword)
	values.Set("client_id", clientId)
	values.Set("client_secret", clientSecret)
	values.Set("username", username)
	values.Set("password", password)

	resp, err := http.PostForm(buildUrl("oauth2/token"), values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, ErrUnauthenticatedRequest
	}

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

	return &Client{
		authenticated: true,
		accessToken:   tresp.AccessToken,
		expiryTime:    time.Now().Add(time.Duration(tresp.ExpiresIn) * time.Second),
	}, nil
}

// ExpiresAt returns the time that the current oauth token expires and will have to be refreshed.
func (c *Client) ExpiresAt() time.Time {
	return c.expiryTime
}

func (c *Client) Authenticated() bool {
	if time.Now().Before(c.ExpiresAt()) {
		return true
	}
	c.authenticated = false
	return c.authenticated
}

// callWithAuth makes authenticated calls to the Mondo API.
func (c *Client) callWithAuth(methodType, URL string, params map[string]string) (*http.Response, error) {
	var resp *http.Response
	var err error

	switch methodType {
	case http.MethodGet:
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

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.accessToken))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			c.authenticated = false
			return nil, ErrUnauthenticatedRequest
		}

	case http.MethodPost:
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}

		req, err := http.NewRequest(methodType, buildUrl(URL), strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.accessToken))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			c.authenticated = false
			return nil, ErrUnauthenticatedRequest
		}
	}

	return resp, err
}

// Transactions returns a slice of Transactions, with the merchant expanded within the Transaction. This endpoint supports pagination. To paginate, provide the last Transacation.ID to the since parameter of the function, if the length of the results that are returned is equal to your limit.
// https://getmondo.co.uk/docs/#list-transactions
func (c *Client) Transactions(accountId, since, before string, limit int) ([]*Transaction, error) {
	params := map[string]string{
		"account_id": accountId,
		"expand[]":   "merchant",
		"limit":      fmt.Sprintf("%v", limit),
		"since":      since,
		"before":     before,
	}

	resp, err := c.callWithAuth(http.MethodGet, "transactions", params)
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
func (c *Client) TransactionByID(accountId, transactionId string) (*Transaction, error) {
	params := map[string]string{
		"account_id": accountId,
		"expand[]":   "merchant",
	}

	resp, err := c.callWithAuth(http.MethodGet, fmt.Sprintf("transactions/%s", transactionId), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrNoTransactionFound
	}

	tresp := &transactionByIDResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &tresp); err != nil {
		return nil, err
	}

	return tresp.Transaction, nil
}

func (c *Client) Accounts() ([]*Account, error) {
	resp, err := c.callWithAuth(http.MethodGet, "accounts", nil)
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

func (c *Client) Balance(accountId string) (*Balance, error) {
	r, err := c.callWithAuth(http.MethodGet, "balance", map[string]string{
		"account_id": accountId,
	})

	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	bal := &Balance{}
	b, err := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(b, bal); err != nil {
		return nil, err
	}

	return bal, nil
}

// CreateFeedItem creates a feed item in the user's application.
func (c *Client) CreateFeedItem(accountId string, item *FeedItem) error {
	if item == nil {
		return errors.New("cannot pass nil item")
	}

	if item.ImageURL == "" {
		return fmt.Errorf("imageURL cannot be empty")
	}

	if accountId == "" {
		return fmt.Errorf("accountId cannot be empty")
	}

	if item.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if item.BgColor == "" {
		item.BgColor = "#FCF1EE"
	}

	if item.BodyColor == "" {
		item.BodyColor = "#FCF1EE"
	}

	if item.TitleColor == "" {
		item.TitleColor = "#333"
	}

	params := map[string]string{
		"account_id":               accountId,
		"type":                     "basic",
		"params[title]":            item.Title,
		"params[image_url]":        item.ImageURL,
		"params[background_color]": item.BgColor,
		"params[body_color]":       item.BodyColor,
		"params[title_color]":      item.TitleColor,
		"params[body]":             item.Body,
	}

	resp, err := c.callWithAuth(http.MethodPost, "feed", params)
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

// Registers a web hook. Each time a matching event occurs, we will make a POST call to the URL you provide. If the call fails, we will retry up to a maximum of 5 attempts, with exponential backoff.
func (c *Client) RegisterWebhook(accountId, URL string) (*Webhook, error) {
	type registerWebhookResponse struct {
		Webhook Webhook `json:"webhook"`
	}

	if accountId == "" {
		return nil, fmt.Errorf("accountId cannot be empty")
	}

	if URL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	params := map[string]string{
		"account_id": accountId,
		"url":        URL,
	}

	resp, err := c.callWithAuth(http.MethodPost, "webhooks", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wresp registerWebhookResponse
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &wresp); err != nil {
		return nil, err
	}

	return &wresp.Webhook, nil
}

// Deletes a web hook. When you delete a web hook, we will no longer send notifications to it.
func (c *Client) DeleteWebhook(webhookId string) error {
	if webhookId == "" {
		return fmt.Errorf("webhookId cannot be empty")
	}

	path := fmt.Sprintf("webhooks/%s", webhookId)
	_, err := c.callWithAuth(http.MethodDelete, path, nil)
	return err
}

// Registers an attachment. Once you have obtained a URL for an attachment, either by uploading to the upload_url obtained from the upload endpoint above or by hosting a remote image, this URL can then be registered against a transaction. Once an attachment is registered against a transaction this will be displayed on the detail page of a transaction within the Mondo app.
func (c *Client) RegisterAttachment(externalId, fileURL, fileType string) (*Attachment, error) {
	if externalId == "" {
		return nil, fmt.Errorf("externalId cannot be empty")
	}

	if fileURL == "" {
		return nil, fmt.Errorf("fileURL cannot be empty")
	}

	if fileType == "" {
		return nil, fmt.Errorf("fileType cannot be empty")
	}

	params := map[string]string{
		"external_id": externalId,
		"file_type":   fileType,
		"file_url":    fileURL,
	}

	resp, err := c.callWithAuth(http.MethodGet, "attachment/register", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var aresp registerAttachmentResponse
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &aresp); err != nil {
		return nil, err
	}

	return &aresp.Attachment, nil
}

func SetBaseEndpoint(address string) {
	baseEndpoint = address
}

func buildUrl(path string) string {
	return fmt.Sprintf("%v/%v", baseEndpoint, path)
}
