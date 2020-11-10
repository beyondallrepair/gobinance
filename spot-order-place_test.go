package gobinance_test

import (
	"context"
	"fmt"
	"github.com/beyondallrepair/gobinance"
	mock_gobinance "github.com/beyondallrepair/gobinance/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type spotOrderTestCase struct {
	name           string
	setup          func(t *testing.T, mocks *clientMocks)
	ctx            context.Context
	options        []gobinance.SpotOrderOption
	call           func(ctx context.Context, uut *gobinance.Client, options []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error)
	errorCheck     errorCheck
	expectedResult gobinance.SpotOrderResult
}

func runSpotOrderTestCases(t *testing.T, testCases ...spotOrderTestCase) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocks := &clientMocks{
				MockDoer:   mock_gobinance.NewMockDoer(ctrl),
				MockSigner: mock_gobinance.NewMockSigner(ctrl),
			}

			u, _ := url.Parse(testBaseURL)
			uut := &gobinance.Client{
				HTTPApiURL: u,
				UserAgent:  testUserAgent,
				APIKey:     testBinanceApiKey,
				RecvWindow: testRecvWindow,
				Signer:     mocks.MockSigner,
				Doer:       mocks.MockDoer,
				Now:        mockNow,
			}

			tc.setup(t, mocks)
			got, err := tc.call(tc.ctx, uut, tc.options)
			if cont := tc.errorCheck(t, err); !cont {
				return
			}

			if diff := cmp.Diff(tc.expectedResult, got, bigFloatComparer); diff != "" {
				t.Errorf("unexpected result.\n%s", diff)
			}
		})
	}
}

func commonSpotTestCases(expectedValues func() url.Values, call func(context.Context, *gobinance.Client, []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error)) []spotOrderTestCase {
	return []spotOrderTestCase{
		{
			name: "request values",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					expectedValues := expectedValues()
					if req.Method != http.MethodPost {
						t.Errorf("unexpected http method: expected %v but got %v", http.MethodPost, req.Method)
					}
					if req.URL.Path != "/api/v3/order" {
						t.Errorf("unexpected path: expected %v but got %v", "/api/v3/order", req.URL.Path)
					}
					if hdr := req.Header.Get("X-MBX-APIKEY"); hdr != testBinanceApiKey {
						t.Errorf("unexpected API key: expected %v but got %v", testBinanceApiKey, hdr)
					}
					if hdr := req.Header.Get("User-Agent"); hdr != testUserAgent {
						t.Errorf("unexpected user agent: expected %v but got %v", testUserAgent, hdr)
					}
					expectedValues.Set("newOrderRespType", "FULL")
					expectedValues.Set("signature", mockSignature)
					expectedValues.Set("timestamp", fmt.Sprint(currentTimeMillis))
					expectedValues.Set("recvWindow", fmt.Sprint(testRecvWindow.Milliseconds()))

					if diff := cmp.Diff(expectedValues, req.URL.Query()); diff != "" {
						t.Errorf("unexpected parameters passed to request:\n%v", diff)
					}
				}).Return(nil, fmt.Errorf("stop early"))
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNotNil,
		},
		{
			name: "http error",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode: 400,
					Body:       ioutil.NopCloser(strings.NewReader(`{ "msg":"test message", "code":-1234 }`)),
				}, nil)
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: isHttpError(400, -1234),
		},
		{
			name: "request error",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(nil, fmt.Errorf("test error"))
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNotNil,
		},
		{
			name: "nil context",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
			},
			ctx:        nil,
			call:       call,
			errorCheck: errNotNil,
		},
		{
			name: "spot client order id",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					if val := req.URL.Query().Get("newClientOrderId"); val != "testOrderID" {
						t.Errorf("unexpected newClientOrderId. expected %v but got %v", "testOrderID", val)
					}
				}).Return(nil, fmt.Errorf("stop early"))
			},
			options: []gobinance.SpotOrderOption{
				gobinance.SpotClientOrderID("testOrderID"),
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNotNil,
		},
		{
			name: "SpotOrderRecvWindow",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					expected := fmt.Sprint((testRecvWindow + time.Second).Milliseconds())
					if got := req.URL.Query().Get("recvWindow"); got != expected {
						t.Errorf("unexpected value for recvWindow. expected %v but got %v", expected, got)
					}
				}).Return(nil, fmt.Errorf("stop early"))
			},
			options: []gobinance.SpotOrderOption{
				gobinance.SpotOrderRecvWindow(testRecvWindow + time.Second),
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNotNil,
		},
		{
			name: "success",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode: 200,
					// from the binance documentation, but with values modified to
					// confirm fields are being correctly assigned
					Body: ioutil.NopCloser(strings.NewReader(`{
					  "symbol": "BTCUSDT",
					  "orderId": 28,
					  "orderListId": -1,
					  "clientOrderId": "6gCrw2kRUAF9CvJDGP16IP",
					  "transactTime": 1507725176595,
					  "price": "1.25000000",
					  "origQty": "2.25000000",
					  "executedQty": "3.25000000",
					  "cummulativeQuoteQty": "4.25000000",
					  "status": "FILLED",
					  "timeInForce": "GTC",
					  "type": "MARKET",
					  "side": "SELL",
					  "fills": [
						{
						  "price": "4000.25000000",
						  "qty": "1.25000000",
						  "commission": "4.25000000",
						  "commissionAsset": "USDT"
						}
					  ]
					}`)),
				}, nil)
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNil,
			expectedResult: gobinance.SpotOrderResult{
				Symbol:             "BTCUSDT",
				OrderID:            28,
				OrderListID:        -1,
				ClientOrderID:      "6gCrw2kRUAF9CvJDGP16IP",
				TransactTime:       time.Date(2017, 10, 11, 12, 32, 56, int(595*time.Millisecond), time.UTC),
				Price:              big.NewFloat(1.25),
				OrigQty:            big.NewFloat(2.25),
				ExecutedQty:        big.NewFloat(3.25),
				CumulativeQuoteQty: big.NewFloat(4.25),
				Status:             gobinance.OrderStatusFilled,
				TimeInForce:        gobinance.TimeInForceGoodTilCanceled,
				Type:               gobinance.OrderTypeMarket, // although this varies per order type, our mock result is always MARKET here
				Side:               gobinance.OrderSideSell,   // Mocked result is SELL
				Fills: []gobinance.Fill{
					{
						Price:           big.NewFloat(4000.25),
						Qty:             big.NewFloat(1.25),
						Commission:      big.NewFloat(4.25),
						CommissionAsset: "USDT",
					},
				},
			},
		},
	}
}

func TestClient_PlaceLimitMakerOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	common := commonSpotTestCases(func() url.Values {
		return map[string][]string{
			"type":     {"LIMIT_MAKER"},
			"symbol":   {testSymbol},
			"side":     {string(orderSide)},
			"quantity": {"1.25"},
			"price":    {"2.25"},
		}
	}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
		// input parameters dont matter here as they're not checked in the common tests
		return client.PlaceLimitMakerOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), big.NewFloat(2.25), opts...)
	})
	runSpotOrderTestCases(t, common...)
}

func TestClient_PlaceLimitOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	common := commonSpotTestCases(func() url.Values {
		return map[string][]string{
			"type":        {"LIMIT"},
			"symbol":      {testSymbol},
			"side":        {string(orderSide)},
			"quantity":    {"1.25"},
			"price":       {"2.25"},
			"timeInForce": {"GTC"},
		}
	}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
		// input parameters dont matter here as they're not checked in the common tests
		return client.PlaceLimitOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), big.NewFloat(2.25), gobinance.TimeInForceGoodTilCanceled, opts...)
	})
	runSpotOrderTestCases(t, common...)
}

func TestClient_PlaceSpotMarketOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	t.Run("QuantityAssetQuote", func(t *testing.T) {
		common := commonSpotTestCases(func() url.Values {
			return map[string][]string{
				"type":          {"MARKET"},
				"symbol":        {testSymbol},
				"side":          {string(orderSide)},
				"quoteOrderQty": {"1.25"},
			}
		}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
			// input parameters dont matter here as they're not checked in the common tests
			return client.PlaceSpotMarketOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), gobinance.QuantityAssetQuote, opts...)
		})
		runSpotOrderTestCases(t, common...)
	})
	t.Run("QuantityAssetBase", func(t *testing.T) {
		common := commonSpotTestCases(func() url.Values {
			return map[string][]string{
				"type":     {"MARKET"},
				"symbol":   {testSymbol},
				"side":     {string(orderSide)},
				"quantity": {"1.25"},
			}
		}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
			// input parameters dont matter here as they're not checked in the common tests
			return client.PlaceSpotMarketOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), gobinance.QuantityAssetBase, opts...)
		})
		runSpotOrderTestCases(t, common...)
	})
}

func TestClient_PlaceStopLossLimitOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	common := commonSpotTestCases(func() url.Values {
		return map[string][]string{
			"type":        {"STOP_LOSS_LIMIT"},
			"symbol":      {testSymbol},
			"side":        {string(orderSide)},
			"quantity":    {"1.25"},
			"stopPrice":   {"2.25"},
			"price":       {"3.25"},
			"timeInForce": {"GTC"},
		}
	}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
		// input parameters dont matter here as they're not checked in the common tests
		return client.PlaceStopLossLimitOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), big.NewFloat(2.25), big.NewFloat(3.25), gobinance.TimeInForceGoodTilCanceled, opts...)
	})
	runSpotOrderTestCases(t, common...)
}

func TestClient_PlaceStopLossOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	common := commonSpotTestCases(func() url.Values {
		return map[string][]string{
			"type":      {"STOP_LOSS"},
			"symbol":    {testSymbol},
			"side":      {string(orderSide)},
			"quantity":  {"1.25"},
			"stopPrice": {"2.25"},
		}
	}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
		// input parameters dont matter here as they're not checked in the common tests
		return client.PlaceStopLossOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), big.NewFloat(2.25), opts...)
	})
	runSpotOrderTestCases(t, common...)
}

func TestClient_PlaceTakeProfitLimitOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	common := commonSpotTestCases(func() url.Values {
		return map[string][]string{
			"type":        {"TAKE_PROFIT_LIMIT"},
			"symbol":      {testSymbol},
			"side":        {string(orderSide)},
			"quantity":    {"1.25"},
			"stopPrice":   {"2.25"},
			"price":       {"3.25"},
			"timeInForce": {"GTC"},
		}
	}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
		// input parameters dont matter here as they're not checked in the common tests
		return client.PlaceTakeProfitLimitOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), big.NewFloat(2.25), big.NewFloat(3.25), gobinance.TimeInForceGoodTilCanceled, opts...)
	})
	runSpotOrderTestCases(t, common...)
}

func TestClient_PlaceTakeProfitOrder(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderSide  = gobinance.OrderSideSell
	)
	common := commonSpotTestCases(func() url.Values {
		return map[string][]string{
			"type":      {"TAKE_PROFIT"},
			"symbol":    {testSymbol},
			"side":      {string(orderSide)},
			"quantity":  {"1.25"},
			"stopPrice": {"2.25"},
		}
	}, func(ctx context.Context, client *gobinance.Client, opts []gobinance.SpotOrderOption) (gobinance.SpotOrderResult, error) {
		// input parameters dont matter here as they're not checked in the common tests
		return client.PlaceTakeProfitOrder(ctx, testSymbol, gobinance.OrderSideSell, big.NewFloat(1.25), big.NewFloat(2.25), opts...)
	})
	runSpotOrderTestCases(t, common...)
}
