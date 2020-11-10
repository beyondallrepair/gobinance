package gobinance

import (
	"context"
	"fmt"
	"net/http"
)

// OpenOrderOptions is a function that applies optional parameters or orverrides to a
// function to query fetch orders
type OpenOrderOptions func(input *openOrderInput)

// OpenSpotOrdersForSymbol fetches the open orders on the specified symbol
func (c *Client) OpenSpotOrdersForSymbol(ctx context.Context, symbol string, opts ...OpenOrderOptions) ([]SpotOrder, error) {
	input := openOrderInput{
		Symbol:  symbol,
	}
	return c.openOrders(ctx, input, opts)
}

// AllOpenSpotOrders fetches all open spot orders on this account, regardless of symbol.
//
// Note that this is an expensive operation, call it sparingly.
func (c *Client) AllOpenSpotOrders(ctx context.Context, opts ...OpenOrderOptions) ([]SpotOrder, error) {
	input := openOrderInput{}
	return c.openOrders(ctx, input, opts)
}

type openOrderInput struct {
	Symbol string `param:"symbol,omitempty"`
}

func applyOpenOrderOptions(input *openOrderInput, opts ...OpenOrderOptions) {
	for _, o := range opts {
		o(input)
	}
}

func (c *Client) openOrders(ctx context.Context, input openOrderInput, opts []OpenOrderOptions) ([]SpotOrder, error) {
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
