package gobinance_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/beyondallrepair/gobinance"
	mock_gobinance "github.com/beyondallrepair/gobinance/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestWebsocketClient_Trades(t *testing.T) {
	t.Parallel()
	const (
		BaseURL           = "wss://example.com"
		TestSymbol        = "testsymbol"
		TestSymbolEncoded = "testsymbol"
	)
	var (
		testError = fmt.Errorf("test error")
	)
	type mocks struct {
		*mock_gobinance.MockDialContexter
		*mock_gobinance.MockNextReaderCloser
	}

	testCases := []struct {
		name         string
		setup        func(context.Context, *mocks) context.Context
		expectations func([]gobinance.TradeEventOrError)
	}{
		{
			name: "dialer errors",
			setup: func(ctx context.Context, mocks *mocks) context.Context{
				mocks.MockDialContexter.EXPECT().DialContext(
					gomock.Not(gomock.Nil()),
					fmt.Sprintf("%s/ws/%s@trade", BaseURL, TestSymbolEncoded),
					nil).Return(nil, nil, testError)
				return ctx
			},
			expectations: func(events []gobinance.TradeEventOrError) {
				if l := len(events); l != 1 {
					t.Errorf("unexpected number of events received. expected 1 but got %v", l)
					return
				}
				if events[0].Err == nil || !errors.Is(events[0].Err, testError) {
					t.Errorf("expected error to extend from\n\t%#v\nbut got\n\t%#v", testError, events[0].Err)
				}
			},
		},
		{
			name: "underlying context is cancelled",
			setup: func(ctx context.Context, mocks *mocks) context.Context {
				mocks.MockDialContexter.EXPECT().DialContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).Return(mocks.MockNextReaderCloser, nil, nil)

				mocks.MockNextReaderCloser.EXPECT().Close() //should always be closed after using
				ctx, cancel := context.WithCancel(ctx)
				mocks.MockNextReaderCloser.EXPECT().NextReader().Do(func() {
					cancel()
					time.Sleep(5 * time.Second)
				})
				return ctx
			},
			expectations: func(events []gobinance.TradeEventOrError) {
				if l := len(events); l != 0 {
					t.Errorf("expected 0 events but got %v", l)
				}
			},
		},
		{
			name: "trade information is received",
			setup: func(ctx context.Context, mocks *mocks) context.Context {
				mocks.MockDialContexter.EXPECT().DialContext(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).Return(mocks.MockNextReaderCloser, nil, nil)
				mocks.MockNextReaderCloser.EXPECT().Close() //should always be closed after using

				// note that the values are manipulated slightly to avoid two events having the same values
				first := mocks.MockNextReaderCloser.EXPECT().NextReader().Return(0,
					bytes.NewBufferString(`{"e":"trade","E":1604705434642,"s":"BTCUSDT","t":455634704,"p":"15617.99000000","q":"0.00720000","b":3530255770,"a":3530255647,"T":1604705434637,"m":false,"M":true}`),
					nil,
				)
				second := mocks.MockNextReaderCloser.EXPECT().NextReader().Return(0,
					bytes.NewBufferString(`{"e":"trade","E":1604705434643,"s":"BTCUSDT","t":455634705,"p":"15617.98000000","q":"0.00200000","b":3530255771,"a":3530255668,"T":1604705434638,"m":true,"M":false}`),
					nil,
				).After(first)
				third := mocks.MockNextReaderCloser.EXPECT().NextReader().Return(0,
					bytes.NewBufferString(`{"e":"trade","E":1604705434644,"s":"BTCUSDT","t":455634706,"p":"15617.97000000","q":"0.05085100","b":3530255772,"a":3530255737,"T":1604705434639,"m":false,"M":true}`),
					nil,
				).After(second)
				mocks.MockNextReaderCloser.EXPECT().NextReader().Return(0, nil, testError).After(third)
				return ctx
			},
			expectations: func(events []gobinance.TradeEventOrError){
				if l := len(events); l != 4 {
					t.Fatalf("expected 4 events but got %v", l)
				}
				// last event should be an error
				if events[3].Err == nil {
					t.Errorf("expected last event to be an error but got %#v", events[3])
				}

				expected := []gobinance.TradeEventOrError{
					{
						TradeEvent: gobinance.TradeEvent{
							Event:         "trade",
							Time:          time.Date(2020,11,06,23,30,34,642*int(time.Millisecond), time.UTC),
							Symbol:        "BTCUSDT",
							TradeID:       455634704,
							Price:         "15617.99000000",
							Quantity:      "0.00720000",
							BuyerOrderID:  3530255770,
							SellerOrderID: 3530255647,
							TradeTime:     time.Date(2020,11,06,23,30,34,637*int(time.Millisecond), time.UTC),
							IsBuyerMaker:  false,
						},
					},
					{
						TradeEvent: gobinance.TradeEvent{
							Event:         "trade",
							Time:          time.Date(2020,11,06,23,30,34,643*int(time.Millisecond), time.UTC),
							Symbol:        "BTCUSDT",
							TradeID:       455634705,
							Price:         "15617.98000000",
							Quantity:      "0.00200000",
							BuyerOrderID:  3530255771,
							SellerOrderID: 3530255668,
							TradeTime:     time.Date(2020,11,06,23,30,34,638*int(time.Millisecond), time.UTC),
							IsBuyerMaker:  true,
						},
					},
					{
						TradeEvent: gobinance.TradeEvent{
							Event:         "trade",
							Time:          time.Date(2020,11,06,23,30,34,644*int(time.Millisecond), time.UTC),
							Symbol:        "BTCUSDT",
							TradeID:       455634706,
							Price:         "15617.97000000",
							Quantity:      "0.05085100",
							BuyerOrderID:  3530255772,
							SellerOrderID: 3530255737,
							TradeTime:     time.Date(2020,11,06,23,30,34,639*int(time.Millisecond), time.UTC),
							IsBuyerMaker:  false,
						},
					},
				}
				if diff := cmp.Diff(expected, events[0:3]); diff != "" {
					t.Errorf("unexpected events.  %s",diff)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDialer := mock_gobinance.NewMockDialContexter(ctrl)
			mockNextReader := mock_gobinance.NewMockNextReaderCloser(ctrl)

			ctx := tc.setup(context.Background(), &mocks{
				MockDialContexter:    mockDialer,
				MockNextReaderCloser: mockNextReader,
			})

			baseURL,_ := url.Parse(BaseURL)
			dialerDialer := 1
			_ = dialerDialer
			uut := &gobinance.Client{
				WebsocketApiURL: baseURL,
				DialContexter:   mockDialer,
			}

			trades := uut.Trades(ctx, TestSymbol)

			var got []gobinance.TradeEventOrError
			for trade := range trades {
				got = append(got, trade)
			}
			tc.expectations(got)
		})
	}
}
