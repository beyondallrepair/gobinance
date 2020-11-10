package gobinance

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
)

// CancelSpotOrderResult holds the data returned by binance in response to a request to cancel an order
type CancelSpotOrderResult struct {
	Symbol                string      `json:"symbol"`
	OriginalClientOrderID string      `json:"origClientOrderId"`
	OrderID               int64       `json:"orderId"`
	OrderListID           int64       `json:"orderListId"`
	ClientOrderID         string      `json:"clientOrderId"`
	Price                 *big.Float  `json:"price"`
	OriginalQty           *big.Float  `json:"origQty"`
	ExecutedQty           *big.Float  `json:"executedQty"`
	CumulativeQuoteQty    *big.Float  `json:"cummulativeQuoteQty"` // note misspelling is intentional
	Status                OrderStatus `json:"status"`
	TimeInForce           TimeInForce `json:"timeInForce"`
	Type                  OrderType   `json:"type"`
	Side                  OrderSide   `json:"side"`
}

// CancelSpotOrderOption is a function that applies optional parameters / overrides to a cancel order operation
type CancelSpotOrderOption func(*cancelSpotOrderInput)

// CancelSpotClientOrderID is a CancelSpotOrderOption which sets the NewClientOrderID value of a request
// to cancel a spot order to `id`.
//
// *Note*: This does not cause an existing order of this ID to be cancelled.  For that, use
// Client.CancelOrderByClientOrderID
func CancelSpotClientOrderID(id string) CancelSpotOrderOption {
	return func(s *cancelSpotOrderInput) {
		s.NewClientOrderID = id
	}
}

type cancelSpotOrderInput struct {
	Symbol            string `param:"symbol,omitempty"`
	OrderID           int64  `param:"orderId,omitempty"`
	OrigClientOrderID string `param:"origClientOrderId,omitempty"`
	NewClientOrderID  string `param:"newClientOrderId,omitempty"`
}

func applyCancelSpotOrderOptions(in *cancelSpotOrderInput, opts ...CancelSpotOrderOption) {
	for _, o := range opts {
		o(in)
	}
}

// CancelOrderByID cancels an order with the given orderID assigned by the exchange when originally
// placing the order.
func (c *Client) CancelOrderByOrderID(ctx context.Context, symbol string, orderID int64, opts ...CancelSpotOrderOption) (CancelSpotOrderResult, error) {
	input := cancelSpotOrderInput{
		Symbol:  symbol,
		OrderID: orderID,
	}
	return c.cancelSpotOrder(ctx, input, opts)
}

// CancelOrderByClientOrderID cancels an order with the given client order ID, a client-generated ID
// optionally passed to the exchage when creating the order
func (c *Client) CancelOrderByClientOrderID(ctx context.Context, symbol string, clientOrderID string, opts ...CancelSpotOrderOption) (CancelSpotOrderResult, error) {
	input := cancelSpotOrderInput{
		Symbol:            symbol,
		OrigClientOrderID: clientOrderID,
	}
	return c.cancelSpotOrder(ctx, input, opts)
}

func (c *Client) cancelSpotOrder(ctx context.Context, input cancelSpotOrderInput, opts []CancelSpotOrderOption) (CancelSpotOrderResult, error) {
	applyCancelSpotOrderOptions(&input, opts...)
	params, err := toURLValues(input)
	if err != nil {
		return CancelSpotOrderResult{}, fmt.Errorf("error building request parmeters: %w", err)
	}
	req, err := c.buildSignedRequest(ctx, http.MethodDelete, "/api/v3/order", params)
	if err != nil {
		return CancelSpotOrderResult{}, fmt.Errorf("error building request: %w", err)
	}

	var out CancelSpotOrderResult
	err = performRequest(c.Doer, req, &out)
	return out, err
}
