package gobinance

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

//go:generate mockgen -destination=mocks/mock_http-client.go . Doer,Signer

// Doer provides a Do method to perform an HTTP request
type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

// Signer provides a method to return an HMAC SHA256 hashed version of the
// input string encoded as a hexadecimal string
type Signer interface {
	Sign(input string) string
}

// NextReaderCloser provides a method for reading messages from a WebSocket
// connection and closing that connection when finished
type NextReaderCloser interface {
	NextReader() (int, io.Reader, error)
	Close() error
}

// DialContexter provides methods for initiating a websocket stream
type DialContexter interface {
	DialContext(ctx context.Context, url string, hdr http.Header) (NextReaderCloser, *http.Response, error)
}

// Client provides methods for interacting with the binance API
type Client struct {
	// HTTPApiURL is the scheme and domain portion of the Binance HTTP API
	// this is usually `https://api.binance.com`
	HTTPApiURL *url.URL
	// WebsocketApiURL is the scheme and domain portion of the Binance Websockets API
	// this is usually `wss://stream.binance.com:9443`
	WebsocketApiURL *url.URL
	// UserAgent is passed in HTTP requests as the `User-Agent` header
	UserAgent string
	// APIKey is the API key to use in authenticated requests to the binance API
	APIKey string
	// RecvWindow is the maximum duration allowed between the client making a request
	// and binance receiving it.  Requests that take longer than this duration are rejected
	// by binance. When this value is 0, binance's default value is used (see their official
	// docs for information)
	RecvWindow time.Duration
	// Signer provides a method for signing requests before sending them to binance
	Signer Signer
	// Doer provides a method for performing HTTP requests
	Doer Doer
	// DialContexter provides a method for making websocket connections
	DialContexter DialContexter
	// Now returns the current time
	Now func() time.Time
}
