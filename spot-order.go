package gobinance

import (
	"encoding/json"
	"math/big"
	"time"
)

// SpotOrder holds the data returned by binance in response to a query order call
type SpotOrder struct {
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

func (r *SpotOrder) UnmarshalJSON(bs []byte) error {
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

	*r = SpotOrder{
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
