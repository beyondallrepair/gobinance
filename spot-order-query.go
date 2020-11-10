package gobinance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// QueryOrderOption is a function that applies optional parameters / overrides to a query order operation
type QueryOrderOption func(input *queryOrderInput)

// QueryOrderResult holds the data returned by binance in response to a query order call
type QueryOrderResult struct {
	Symbol             string
	OrderID            int64
	OrderListID        int64
	ClientOrderID      string
	Price              *big.Float
	OriginalQty        *big.Float
	ExecutedQty        *big.Float
	CumulativeQuoteQty    *big.Float
	Status                OrderStatus
	TimeInForce           TimeInForce
	Type                  OrderType
	Side                  OrderSide
	StopPrice             *big.Float
	IcebergQty            *big.Float
	Time                  time.Time
	UpdateTime            time.Time
	IsWorking             bool
	OriginalQuoteOrderQty *big.Float
}

func (r *QueryOrderResult) UnmarshalJSON(bs []byte) error {
	var tmp struct {
		Symbol             string          `json:"symbol"`
		OrderID            int64           `json:"orderId"`
		OrderListID        int64           `json:"orderListId"`
		ClientOrderID      string          `json:"clientOrderId"`
		Price              *big.Float      `json:"price"`
		OriginalQty        *big.Float      `json:"origQty"`
		ExecutedQty        *big.Float      `json:"executedQty"`
		CumulativeQuoteQty *big.Float      `json:"cummulativeQuoteQty"` // misspelling intentional
		Status             OrderStatus     `json:"status"`
		TimeInForce        TimeInForce     `json:"timeInForce"`
		Type               OrderType       `json:"type"`
		Side               OrderSide       `json:"side"`
		StopPrice          *big.Float      `json:"stopPrice"`
		IcebergQty         *big.Float      `json:"icebergQty"`
		Time               millisTimestamp `json:"time"`
		UpdateTime         millisTimestamp `json:"updateTime"`
		IsWorking          bool            `json:"isWorking"`
		OrigQuoteOrderQty  *big.Float      `json:"origQuoteOrderQty"`
	}
	if err := json.Unmarshal(bs, &tmp); err != nil {
		return err
	}

	*r = QueryOrderResult{
		Symbol:             tmp.Symbol,
		OrderID:            tmp.OrderID,
		OrderListID:        tmp.OrderListID,
		ClientOrderID:      tmp.ClientOrderID,
		Price:              tmp.Price,
		OriginalQty:        tmp.OriginalQty,
		ExecutedQty:        tmp.ExecutedQty,
		CumulativeQuoteQty:    tmp.CumulativeQuoteQty,
		Status:                tmp.Status,
		TimeInForce:           tmp.TimeInForce,
		Type:                  tmp.Type,
		Side:                  tmp.Side,
		StopPrice:             tmp.StopPrice,
		IcebergQty:            tmp.IcebergQty,
		Time:                  time.Time(tmp.Time),
		UpdateTime:            time.Time(tmp.UpdateTime),
		IsWorking:             tmp.IsWorking,
		OriginalQuoteOrderQty: tmp.OrigQuoteOrderQty,
	}
	return nil
}

// QueryOrderByID fetches the order whose ID as assigned by the exchange is orderID
func (c *Client) QueryOrderByID(ctx context.Context, symbol string, orderID int64, opts ...QueryOrderOption) (QueryOrderResult, error) {
	input := queryOrderInput{
		Symbol:  symbol,
		OrderID: orderID,
	}
	return c.queryOrder(ctx, input, opts)
}

// QueryOrderByClientID fetches the order whose ID as assigned by the client is clientOrderID
func (c *Client) QueryOrderByClientID(ctx context.Context, symbol string, clientOrderID string, opts ...QueryOrderOption) (QueryOrderResult, error) {
	input := queryOrderInput{
		Symbol:            symbol,
		OrigClientOrderID: clientOrderID,
	}
	return c.queryOrder(ctx, input, opts)
}

func (c *Client) queryOrder(ctx context.Context, input queryOrderInput, opts []QueryOrderOption) (QueryOrderResult, error) {
	applyQueryOrderOptions(&input, opts)
	params, err := toURLValues(input)
	if err != nil {
		return QueryOrderResult{}, fmt.Errorf("error building request parameters: %w", err)
	}
	req, err := c.buildSignedRequest(ctx, http.MethodGet, "/api/v3/order", params)
	if err != nil {
		return QueryOrderResult{}, fmt.Errorf("error building request: %w", err)
	}

	var out QueryOrderResult
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
}
