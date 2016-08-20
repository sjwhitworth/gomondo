package mondo

import "time"

type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ClientID     string `json:"client_id"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	Error        string `json:"error"`
}

type Account struct {
	ID            string    `json:"id"`
	AccountNumber string    `json:"account_number"`
	SortCode      string    `json:"sort_code"`
	Description   string    `json:"description"`
	Created       time.Time `json:"created"`
}

type Transaction struct {
	AccountBalance int                    `json:"account_balance"`
	Amount         int                    `json:"amount"`
	Attachments    []interface{}          `json:"attachments"`
	Category       string                 `json:"category"`
	Created        time.Time              `json:"created"`
	Currency       string                 `json:"currency"`
	Description    string                 `json:"description"`
	ID             string                 `json:"id"`
	IsLoad         bool                   `json:"is_load"`
	Merchant       Merchant               `json:"merchant"`
	Metadata       map[string]interface{} `json:"metadata"`
	Notes          string                 `json:"notes"`
	Settled        string                 `json:"settled"`
}

type Merchant struct {
	Address  MerchantAddress `json:"address"`
	Category string          `json:"category"`
	Created  string          `json:"created"`
	Emoji    string          `json:"emoji"`
	GroupID  string          `json:"group_id"`
	ID       string          `json:"id"`
	Logo     string          `json:"logo"`
	Name     string          `json:"name"`
	Online   bool            `json:"online"`
}

type MerchantAddress struct {
	Address        string  `json:"address"`
	Approximate    bool    `json:"approximate"`
	City           string  `json:"city"`
	Country        string  `json:"country"`
	Formatted      string  `json:"formatted"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Postcode       string  `json:"postcode"`
	Region         string  `json:"region"`
	ShortFormatted string  `json:"short_formatted"`
	ZoomLevel      int     `json:"zoom_level"`
}

type Webhook struct {
	AccountId string `json:"account_id"`
	Id        string `json:"id"`
	Url       string `json:"url"`
}

type Attachment struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	ExternalId string `json:"external_id"`
	FileUrl    string `json:"file_url"`
	FileType   string `json:"file_type"`
	Created    string `json:"created"`
}

type WebhookRequest struct {
	Type string       `json:"type"`
	Data *Transaction `json:"data"`
}

type FeedItem struct {
	Title      string
	ImageURL   string
	BgColor    string
	BodyColor  string
	TitleColor string
	Body       string
}

type Balance struct {
	Balance    int64  `json:"balance"`
	Currency   string `json:"currency"`
	SpendToday int64  `json:"spend_today"`
}
