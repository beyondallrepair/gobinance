package gobinance

import "fmt"

// OrderStatus is an enumeration of possible order statuses
type OrderStatus string

const (
	// OrderStatusNew indicates the order has been accepted by the engine
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusPartiallyFilled indicates a part of the order has been filled.
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	// OrderStatusFilled indicates the order has been completed.
	OrderStatusFilled OrderStatus = "FILLED"
	// OrderStatusCancelled indicates the order has been canceled by the user.
	OrderStatusCancelled OrderStatus = "CANCELLED"
	// OrderStatusRejected indicates the order was not accepted by the engine and not processed.
	OrderStatusRejected OrderStatus = "REJECTED"
	// OrderStatusExpired indicates the order was canceled according to the order type's rules (e.g. LIMIT FOK orders with no
	// fill, LIMIT IOC or MARKET orders that partially fill) or by the exchange, (e.g. orders canceled during
	// liquidation, orders canceled during maintenance)
	OrderStatusExpired OrderStatus = "EXPIRED"
)

// Validate returns nil if the value is a valid OrderStatus, or an error if not.
func (s OrderStatus) Validate() error {
	switch s {
	case OrderStatusNew:
	case OrderStatusPartiallyFilled:
	case OrderStatusFilled:
	case OrderStatusCancelled:
	case OrderStatusRejected:
	case OrderStatusExpired:
	default:
		return fmt.Errorf("OrderStatus, %q, is not known", s)
	}
	return nil
}

// OrderType is an enumeration of the possible types of orders that can be placed
type OrderType string

const (
	// OrderTypeLimit indicates a limit order
	OrderTypeLimit OrderType = "LIMIT"
	// OrderTypeMarket indicates a market order
	OrderTypeMarket OrderType = "MARKET"
	// OrderTypeStopLoss indicates a stop loss order
	OrderTypeStopLoss OrderType = "STOP_LOSS"
	// OrderTypeStopLossLimit indicates a stop loss limit order
	OrderTypeStopLossLimit OrderType = "STOP_LOSS_LIMIT"
	// OrderTypeTakeProfit indicates a take profit order
	OrderTypeTakeProfit OrderType = "TAKE_PROFIT"
	// OrderTypeTakeProfitLimit indicates a take profit limit order
	OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"
	// OrderTypeLimitMaker indicates a limit maker order
	OrderTypeLimitMaker OrderType = "LIMIT_MAKER"
)

// Validate returns nil if the value is a valid OrderType, or an error if not.
func (t OrderType) Validate() error {
	switch t {
	case OrderTypeLimit:
	case OrderTypeMarket:
	case OrderTypeStopLoss:
	case OrderTypeStopLossLimit:
	case OrderTypeTakeProfit:
	case OrderTypeTakeProfitLimit:
	case OrderTypeLimitMaker:
	default:
		return fmt.Errorf("OrderType, %q, is not known", t)
	}
	return nil
}

// OrderResponseType is an enumeration of possible responses that can be returned for new orders
type OrderResponseType string

const (
	OrderResponseTypeAck    OrderResponseType = "ACK"
	OrderResponseTypeResult OrderResponseType = "RESULT"
	OrderResponseTypeFull   OrderResponseType = "FULL"
)

// Validate returns nil if the value is a valid OrderResponseType, or an error if not.
func (t OrderResponseType) Validate() error {
	switch t {
	case OrderResponseTypeAck:
	case OrderResponseTypeResult:
	case OrderResponseTypeFull:
	default:
		return fmt.Errorf("OrderResponseType, %q, is not known", t)
	}
	return nil
}

// OrderSide indicates the direction of a trade
type OrderSide string

const (
	// OrderSideBuy indicates the order is a 'buy'
	OrderSideBuy OrderSide = "BUY"
	// OrderSideSell indicates the order is a 'sell'
	OrderSideSell OrderSide = "SELL"
)

// Validate returns nil if the value is a valid OrderSide, or an error if not.
func (e OrderSide) Validate() error {
	switch e {
	case OrderSideBuy:
	case OrderSideSell:
	default:
		return fmt.Errorf("OrderSide, %q, is not known", e)
	}
	return nil
}

// TimeInForce is an enumeration of possible expiration behaviours of an order
type TimeInForce string

const (
	// TimeInForceGoodTilCancelled indicates an order will be on the book unless the order is canceled
	TimeInForceGoodTilCancelled TimeInForce = "GTC"
	// TimeInForceImmediateOrCancel indicates an order will try to fill the order as much as it can before the order expires
	TimeInForceImmediateOrCancel TimeInForce = "IOC"
	// TimeInForceFillOrKill indicates an order will expire if the full order cannot be filled upon execution
	TimeInForceFillOrKill TimeInForce = "FOK"
)

// Validate returns nil if the value is a valid TimeInForce, or an error if not.
func (e TimeInForce) Validate() error {
	switch e {
	case TimeInForceGoodTilCancelled:
	case TimeInForceImmediateOrCancel:
	case TimeInForceFillOrKill:
	default:
		return fmt.Errorf("TimeInForce, %q, is not known", e)
	}
	return nil
}

// QuantityAsset is an enumeration of possible assets a quantity is related to in a trade
type QuantityAsset string

const (
	// QuantityAssetQuote indicates the quantity relates to the quote asset
	QuantityAssetQuote = "QUOTE"
	// QuantityAssetBase indicates the quantity relates to the base asset
	QuantityAssetBase = "BASE"
)
