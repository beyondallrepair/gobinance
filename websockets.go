package gobinance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:generate mockgen -destination=mocks/mock_websockets.go . NextReaderCloser,DialContexter
// NextReaderCloser provides methods to read the next message in a websocket stream and to close that stream
type NextReaderCloser interface {
	NextReader() (int, io.Reader, error)
	Close() error
}

// DialContexter provides methods for initiating a websocket stream
type DialContexter interface {
	DialContext(ctx context.Context, url string, hdr http.Header) (NextReaderCloser, *http.Response, error)
}

// WebsocketClient is a client for the binance websockets API
type WebsocketClient struct {
	BaseURL       *url.URL
	DialContexter DialContexter
}

type readerError struct {
	io.Reader
	error
}

// TradeEvent define websocket trade event
type TradeEvent struct {
	Event         string
	Time          time.Time
	Symbol        string
	TradeID       int64
	Price         string
	Quantity      string
	BuyerOrderID  int64
	SellerOrderID int64
	TradeTime     time.Time
	IsBuyerMaker  bool
}

// UnmarshalJSON provides custom unmarshalling for TradeEvents.
func (t *TradeEvent) UnmarshalJSON(bs []byte) error {
	var tmp struct {
		Event         string `json:"e"`
		Time          int64  `json:"E"`
		Symbol        string `json:"s"`
		TradeID       int64  `json:"t"`
		Price         string `json:"p"`
		Quantity      string `json:"q"`
		BuyerOrderID  int64  `json:"b"`
		SellerOrderID int64  `json:"a"`
		TradeTime     int64  `json:"T"`
		IsBuyerMaker  bool   `json:"m"`
		Placeholder   bool   `json:"M"` // add this field to avoid case insensitive unmarshaling
	}
	if err := json.Unmarshal(bs, &tmp); err != nil {
		return err
	}
	*t = TradeEvent{
		Event:         tmp.Event,
		Time:          time.Unix(0, tmp.Time*int64(time.Millisecond)).UTC(),
		Symbol:        tmp.Symbol,
		TradeID:       tmp.TradeID,
		Price:         tmp.Price,
		Quantity:      tmp.Quantity,
		BuyerOrderID:  tmp.BuyerOrderID,
		SellerOrderID: tmp.SellerOrderID,
		TradeTime:     time.Unix(0, tmp.TradeTime*int64(time.Millisecond)).UTC(),
		IsBuyerMaker:  tmp.IsBuyerMaker,
	}
	return nil
}

// TradeEventOrError is a union of TradeEvent or error
type TradeEventOrError struct {
	TradeEvent
	Err error
}

// Trades initiates a websocket connection to binance and returns a channel from which live trades can be streamed from
// binance.  The channel is closed when the underlying context is cancelled, or  upon a connection error or the server
// closing the connection.
func (c *WebsocketClient) Trades(ctx context.Context, symbol string) <-chan TradeEventOrError {
	out := make(chan TradeEventOrError, 1)
	handle := func(reader io.Reader, err error) {
		if err != nil {
			out <- TradeEventOrError{Err: err}
			return
		}
		var trade TradeEvent
		dec := json.NewDecoder(reader)
		if err := dec.Decode(&trade); err != nil {
			out <- TradeEventOrError{Err: fmt.Errorf("error decoding trade event: %w", err)}
			return
		}
		out <- TradeEventOrError{TradeEvent: trade}
	}
	path := fmt.Sprintf("/ws/%s@trade", url.PathEscape(strings.ToLower(symbol)))

	go c.openWebsocket(ctx, path, handle, func() {
		close(out)
	})
	return out
}

// openWebsocket does some generic handling of websocket streams.  It initiates a connection to the endpoint
// using the BaseURL from WebsocketClient and the path provided as an input parameter.  For each event streamed
// from the websocket, the `handler` is called.
//
// This function blocks until the websocket stream is closed either from the server, or due to the underlying
// context being cancelled or a connection error.
func (c *WebsocketClient) openWebsocket(ctx context.Context, path string, handle func(reader io.Reader, err error), after func()) {
	defer after()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	u := c.BaseURL.ResolveReference(&url.URL{
		Path: path,
	})
	con, _, err := c.DialContexter.DialContext(ctx, u.String(), nil)
	if err != nil {
		handle(nil, fmt.Errorf("unable to establish websocket connection: %w", err))
		return
	}
	defer con.Close()

	wsMessages := make(chan readerError, 0)
	go func() {
		defer close(wsMessages)
		for {
			_, msg, err := con.NextReader()
			if err == nil {
				// note that we need to make a copy of the buffer here to avoid
				// races with the consumer vs this loop's next iteration
				var buf []byte
				buf, err = ioutil.ReadAll(msg)
				msg = bytes.NewBuffer(buf)
			}
			select {
			case <-ctx.Done():
				return
			case wsMessages <- readerError{msg, err}:
				if err != nil {
					// errors are permanent, so break the loop
					return
				}
			}
		}
	}()

	for {
		select {
		case msg, ok := <-wsMessages:
			if !ok {
				return
			}
			handle(msg.Reader, msg.error)
		case <-ctx.Done():
			return
		}
	}
}
