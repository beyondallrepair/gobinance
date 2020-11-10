package gobinance

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// QueryOrderOption is a function that applies optional parameters / overrides to a query order operation
type QueryOrderOption func(input *queryOrderInput)

// QueryOrderRecvWindow overrides the default receive window for a query order operation
func QueryOrderRecvWindow (d time.Duration) QueryOrderOption {
	return func(input *queryOrderInput) {
		input.RecvWindow = d.Milliseconds()
	}
}

// QueryOrderByID fetches the order whose ID as assigned by the exchange is orderID
func (c *Client) QueryOrderByID(ctx context.Context, symbol string, orderID int64, opts ...QueryOrderOption) (SpotOrder, error) {
	input := queryOrderInput{
		Symbol:  symbol,
		OrderID: orderID,
	}
	return c.queryOrder(ctx, input, opts)
}

// QueryOrderByClientID fetches the order whose ID as assigned by the client is clientOrderID
func (c *Client) QueryOrderByClientID(ctx context.Context, symbol string, clientOrderID string, opts ...QueryOrderOption) (SpotOrder, error) {
	input := queryOrderInput{
		Symbol:            symbol,
		OrigClientOrderID: clientOrderID,
	}
	return c.queryOrder(ctx, input, opts)
}

func (c *Client) queryOrder(ctx context.Context, input queryOrderInput, opts []QueryOrderOption) (SpotOrder, error) {
	applyQueryOrderOptions(&input, opts)
	params, err := toURLValues(input)
	if err != nil {
		return SpotOrder{}, fmt.Errorf("error building request parameters: %w", err)
	}
	req, err := c.buildSignedRequest(ctx, http.MethodGet, "/api/v3/order", params)
	if err != nil {
		return SpotOrder{}, fmt.Errorf("error building request: %w", err)
	}

	var out SpotOrder
	err = performRequest(c.Doer, req, &out)
	return out, err
}

func applyQueryOrderOptions(in *queryOrderInput, opts []QueryOrderOption) {
	for _, o := range opts {
		o(in)
	}
}

type queryOrderInput struct {
	Symbol            string `param:"symbol"`
	OrderID           int64  `param:"orderId,omitempty"`
	OrigClientOrderID string `param:"origClientOrderId,omitempty"`
	RecvWindow        int64  `param:"recvWindow,omitempty"`
}
