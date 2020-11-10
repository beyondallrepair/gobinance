package gobinance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type spotOrderInput struct {
	Symbol           string            `param:"symbol,omitempty"`
	Side             OrderSide         `param:"side,omitempty"`
	Type             OrderType         `param:"type,omitempty"`
	TimeInForce      TimeInForce       `param:"timeInForce,omitempty"`
	Quantity         *big.Float        `param:"quantity,omitempty"`
	QuoteOrderQty    *big.Float        `param:"quoteOrderQty,omitempty"`
	Price            *big.Float        `param:"price,omitempty"`
	NewClientOrderID string            `param:"newClientOrderId,omitempty"`
	StopPrice        *big.Float        `param:"stopPrice,omitempty"`
	IcebergQty       int               `param:"icebergQty,omitempty"`
	NewOrderRespType OrderResponseType `param:"newOrderRespType,omitempty"`
	RecvWindow       int64             `param:"recvWindow,omitempty"`
}

type Fill struct {
	Price           *big.Float
	Qty             *big.Float
	Commission      *big.Float
	CommissionAsset string
}

type SpotOrderResult struct {
	Symbol             string
	OrderID            int
	OrderListID        int
	ClientOrderID      string
	TransactTime       time.Time
	Price              *big.Float
	OrigQty            *big.Float
	ExecutedQty        *big.Float
	CumulativeQuoteQty *big.Float
	Status             OrderStatus
	TimeInForce        TimeInForce
	Type               OrderType
	Side               OrderSide
	Fills              []Fill
}

func (s *SpotOrderResult) UnmarshalJSON(bs []byte) error {
	var tmp struct {
		Symbol              string
		OrderID             int
		OrderListID         int
		ClientOrderID       string
		TransactTime        millisTimestamp
		Price               *big.Float
		OrigQty             *big.Float
		ExecutedQty         *big.Float
		CummulativeQuoteQty *big.Float // note that the spelling error is intentional -- it is spelt that way in the API
		Status              OrderStatus
		TimeInForce         TimeInForce
		Type                OrderType
		Side                OrderSide
		Fills               []Fill
	}
	if err := json.Unmarshal(bs, &tmp); err != nil {
		return err
	}
	*s = SpotOrderResult{
		Symbol:             tmp.Symbol,
		OrderID:            tmp.OrderID,
		OrderListID:        tmp.OrderListID,
		ClientOrderID:      tmp.ClientOrderID,
		TransactTime:       time.Time(tmp.TransactTime),
		Price:              tmp.Price,
		OrigQty:            tmp.OrigQty,
		ExecutedQty:        tmp.ExecutedQty,
		CumulativeQuoteQty: tmp.CummulativeQuoteQty,
		Status:             tmp.Status,
		TimeInForce:        tmp.TimeInForce,
		Type:               tmp.Type,
		Side:               tmp.Side,
		Fills:              tmp.Fills,
	}
	return nil
}

// SpotOrderOption is a function that applies optional values / overrides to a spot order
type SpotOrderOption func(s *spotOrderInput)

// SpotClientOrderID is a SpotOrderOption which sets the NewClientOrderID value of a spot
// order to `id`.
func SpotClientOrderID(id string) SpotOrderOption {
	return func(s *spotOrderInput) {
		s.NewClientOrderID = id
	}
}

// SpotOrderRecvWindow overrides the default receive window for a request to place a spot order
func SpotOrderRecvWindow(d time.Duration) SpotOrderOption {
	return func(s *spotOrderInput) {
		s.RecvWindow = d.Milliseconds()
	}
}

func applySpotOrderOptions(input *spotOrderInput, opts ...SpotOrderOption) {
	for _, o := range opts {
		o(input)
	}
}

// PlaceLimitOrder places limit order on the spot market.
func (c *Client) PlaceLimitOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, price *big.Float, tif TimeInForce, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:        OrderTypeLimit,
		Symbol:      symbol,
		Side:        side,
		Quantity:    qty,
		Price:       price,
		TimeInForce: tif,
	}

	return c.placeOrder(ctx, input, opts)
}

// PlaceSpotMarketOrder places a market order on the spot market.
//
// When asset is QuantityAssetBase, the qty parameter specifies the amount of the base asset to buy or sell at the market price.
// When asset is QuantityAssetQuote, the qty parameter specifies the amount you'd like to spend or earn, at the current market price.
// The quantity of the base asset is then calculated based on the current market price.
func (c *Client) PlaceSpotMarketOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, asset QuantityAsset, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:   OrderTypeMarket,
		Symbol: symbol,
		Side:   side,
	}
	switch asset {
	case QuantityAssetBase:
		input.Quantity = qty
	case QuantityAssetQuote:
		input.QuoteOrderQty = qty
	default:
		return SpotOrderResult{}, fmt.Errorf("unknown asset value '%v'", asset)
	}
	return c.placeOrder(ctx, input, opts)
}

// PlaceStopLossOrder places a stop-loss order on the spot market.
//
// The order will execute once the `stopPrice` is reached.
func (c *Client) PlaceStopLossOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, stopPrice *big.Float, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:      OrderTypeStopLoss,
		Symbol:    symbol,
		Side:      side,
		Quantity:  qty,
		StopPrice: stopPrice,
	}
	return c.placeOrder(ctx, input, opts)
}

// PlaceStopLossLimitOrder places a stop-loss-limit order on the spot market.
//
// A limit order will be placed once the stop price is reached.
func (c *Client) PlaceStopLossLimitOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, stopPrice *big.Float, limitPrice *big.Float, tif TimeInForce, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:        OrderTypeStopLossLimit,
		Symbol:      symbol,
		Side:        side,
		Quantity:    qty,
		StopPrice:   stopPrice,
		Price:       limitPrice,
		TimeInForce: tif,
	}
	return c.placeOrder(ctx, input, opts)
}

// PlaceTakeProfitOrder places a take-profit order on the spot market.
//
// The order will execute once the stopPrice is reached.
func (c *Client) PlaceTakeProfitOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, stopPrice *big.Float, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:      OrderTypeTakeProfit,
		Symbol:    symbol,
		Side:      side,
		Quantity:  qty,
		StopPrice: stopPrice,
	}
	return c.placeOrder(ctx, input, opts)
}

// PlaceTakeProfitLimitOrder places a take-profit-limit order on the spot market.
//
// A limit order will be placed once the stop price is reached.
func (c *Client) PlaceTakeProfitLimitOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, stopPrice *big.Float, limitPrice *big.Float, tif TimeInForce, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:        OrderTypeTakeProfitLimit,
		Symbol:      symbol,
		Side:        side,
		Quantity:    qty,
		StopPrice:   stopPrice,
		Price:       limitPrice,
		TimeInForce: tif,
	}
	return c.placeOrder(ctx, input, opts)
}

// PlaceLimitMakerOrder places a limit order on the spot market which is immediately rejected if
// it would not be executed as a maker order.
//
// AKA a post-only order.
func (c *Client) PlaceLimitMakerOrder(ctx context.Context, symbol string, side OrderSide, qty *big.Float, price *big.Float, opts ...SpotOrderOption) (SpotOrderResult, error) {
	input := spotOrderInput{
		Type:     OrderTypeLimitMaker,
		Symbol:   symbol,
		Side:     side,
		Quantity: qty,
		Price:    price,
	}
	return c.placeOrder(ctx, input, opts)
}

func (c *Client) placeOrder(ctx context.Context, input spotOrderInput, opts []SpotOrderOption) (SpotOrderResult, error) {
	input.NewOrderRespType = OrderResponseTypeFull
	applySpotOrderOptions(&input, opts...)
	params, err := toURLValues(input)
	if err != nil {
		return SpotOrderResult{}, fmt.Errorf("error building request parameters: %w", err)
	}

	req, err := c.buildSignedRequest(ctx, http.MethodPost, "/api/v3/order", params)
	if err != nil {
		return SpotOrderResult{}, fmt.Errorf("error building request: %w", err)
	}

	var result SpotOrderResult
	err = performRequest(c.Doer, req, &result)
	return result, err
}
