# go-mondo

Package go-mondo provides Go bindings for the Mondo banking app and marshals them into native data structures with full support for Mondo objects like Transactions, Merchants and Addresses. It makes no assumptions regarding your use case. Thus, strategies for pagination, retries and refreshing oauth tokens are left for the caller to decide upon. The full documentation for the API is available [here.](https://getmondo.co.uk/docs)

![pDpOAr](http://cdn.makeagif.com/media/11-29-2015/pDpOAr.gif)

## Supported

* OAuth2 authentication
* Listing accounts
* Reading all transactions
* Reading a specific transaction
* Creating a feed item in your feed, with full styling

## Example

```go
client, err := mondo.Authenticate(clientId, clientSecret, userName, password)
if err != nil {
  return err
}

log.Infof("Authenticated with Mondo successfully!")

// Retrieve all of the accounts.
acs, err := client.Accounts()
if err != nil {
  return err
}

// Grab our account ID.
accountId := acs[0].ID

// Get all transactions. You can also get a specific transaction by ID.
transactions, err := client.Transactions(accountId, "", "", 100)
if err != nil {
  return err
}
log.Infof("%#v", transactions)

// Create new feed item
err := client.CreateFeedItem(accountId, "Morning!", "https://blog.golang.org/gopher/gopher.png", "", "", "", "Hi from go-mondo!")
if err != nil {
  return err
}
```

A larger example of how to use the client is provided in the bankterm example. It takes your last 100 Mondo transactions and prints them to a table in your terminal.

## Things still to do

* Greater client test coverage


## Helping out
Please do.
