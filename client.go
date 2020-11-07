package gobinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

//go:generate mockgen -destination=mocks/mock_http-client.go . Doer,Signer

// Doer provides a Do method to perform an HTTP request
type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Signer interface {
	Sign(input string) string
}

const (
	signatureQuery  = "signature"
	timestampQuery  = "timestamp"
	recvWindowQuery = "recvWindow"
)

type HttpClient struct {
	BaseURL    *url.URL
	UserAgent  string
	APIKey     string
	RecvWindow time.Duration
	Signer     Signer
	Doer       Doer
	Now        func() time.Time
}

// AccountInformation fetches data about the account associated with the API Key provided to
// the client, or an error in the event of connection issues or a non-200 response from binance.
func (c *HttpClient) AccountInformation(ctx context.Context) (AccountInformation, error) {
	req, err := c.buildSignedRequest(ctx, http.MethodGet, "/api/v3/account", nil)
	if err != nil {
		return AccountInformation{}, fmt.Errorf("error building request: %w", err)
	}
	var out AccountInformation
	if err := performRequest(c.Doer, req, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *HttpClient) buildUnsignedRequest(ctx context.Context, method string, path string, parameters url.Values, includeAPIKey bool) (*http.Request, error) {
	u := c.BaseURL.ResolveReference(&url.URL{
		Path:     path,
		RawQuery: parameters.Encode(),
	})

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error building request")
	}

	req.Header.Set("User-Agent", c.UserAgent)
	if includeAPIKey {
		req.Header.Set("X-MBX-APIKEY", c.APIKey)
	}

	return req, nil
}

func (c *HttpClient) buildSignedRequest(ctx context.Context, method string, path string, parameters url.Values) (*http.Request, error) {
	if parameters == nil {
		parameters = make(url.Values)
	}
	parameters.Set(timestampQuery, fmt.Sprint(c.Now().UnixNano()/int64(time.Millisecond)))
	if parameters.Get(recvWindowQuery) == "" && c.RecvWindow > 0 {
		parameters.Set(recvWindowQuery, fmt.Sprint(c.RecvWindow.Milliseconds()))
	}
	signature := c.Signer.Sign(parameters.Encode())
	parameters.Set(signatureQuery, signature)

	return c.buildUnsignedRequest(ctx, method, path, parameters, true)
}

// performRequest executes the request in req using the doer.  If the host returns a non-200 status, an `HttpError`
// error is returned.  If the request succeeds, and `response` is not nil, the JSON body in the response is decoded
// into `response`.
func performRequest(doer Doer, req *http.Request, response interface{}) error {
	resp, err := doer.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errorBody errorDTO
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&errorBody); err != nil {
			return &HttpError{
				HttpStatus: resp.StatusCode,
			}
		}
		return &HttpError{
			HttpStatus: resp.StatusCode,
			errorDTO:   errorBody,
		}
	}

	if response == nil {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(response); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}
	return nil
}
