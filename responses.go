package mondo

type transactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
}

type transactionByIDResponse struct {
	Transaction *Transaction `json:"transaction"`
}

type accountsResponse struct {
	Accounts []*Account `json:"accounts"`
}

type feedItemResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type registerAttachmentResponse struct {
	Attachment Attachment `json:"attachment"`
}
