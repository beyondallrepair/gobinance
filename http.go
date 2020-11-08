package gobinance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	signatureQuery  = "signature"
	timestampQuery  = "timestamp"
	recvWindowQuery = "recvWindow"
	userAgentHeader = "User-Agent"
	apiKeyHeader    = "X-MBX-APIKEY"
)

func (c *Client) buildUnsignedRequest(ctx context.Context, method string, path string, parameters url.Values, includeAPIKey bool) (*http.Request, error) {
	u := c.HTTPApiURL.ResolveReference(&url.URL{
		Path:     path,
		RawQuery: parameters.Encode(),
	})

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error building request")
	}

	req.Header.Set(userAgentHeader, c.UserAgent)
	if includeAPIKey {
		req.Header.Set(apiKeyHeader, c.APIKey)
	}

	return req, nil
}

func (c *Client) buildSignedRequest(ctx context.Context, method string, path string, parameters url.Values) (*http.Request, error) {
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
