package gobinance

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// OpenOrdersOptions is a function that applies optional parameters or overrides to a
// function to query fetch orders
type OpenOrdersOptions func(input *openOrderInput)

// OpenOrdersRecvWindow overrides the receive window for an OpenOrder call.
// See Client.ReceiveWindow for more information
func OpenOrdersRecvWindow(d time.Duration) OpenOrdersOptions {
	return func(input *openOrderInput) {
		input.RecvWindow = d.Milliseconds()
	}
}

// OpenSpotOrdersForSymbol fetches the open orders on the specified symbol
func (c *Client) OpenSpotOrdersForSymbol(ctx context.Context, symbol string, opts ...OpenOrdersOptions) ([]SpotOrder, error) {
	input := openOrderInput{
		Symbol: symbol,
	}
	return c.openOrders(ctx, input, opts)
}

// AllOpenSpotOrders fetches all open spot orders on this account, regardless of symbol.
//
// Note that this is an expensive operation, call it sparingly.
func (c *Client) AllOpenSpotOrders(ctx context.Context, opts ...OpenOrdersOptions) ([]SpotOrder, error) {
	input := openOrderInput{}
	return c.openOrders(ctx, input, opts)
}

type openOrderInput struct {
	Symbol     string `param:"symbol,omitempty"`
	RecvWindow int64  `param:"recvWindow,omitempty"`
}

func applyOpenOrderOptions(input *openOrderInput, opts ...OpenOrdersOptions) {
	for _, o := range opts {
		o(input)
	}
}

func (c *Client) openOrders(ctx context.Context, input openOrderInput, opts []OpenOrdersOptions) ([]SpotOrder, error) {
	applyOpenOrderOptions(&input, opts...)
	params, err := toURLValues(input)
	if err != nil {
		return nil, fmt.Errorf("error building request parameters: %w", err)
	}

	req, err := c.buildSignedRequest(ctx, http.MethodGet, "/api/v3/openOrders", params)
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	var result []SpotOrder
	err = performRequest(c.Doer, req, &result)
	return result, err
}
