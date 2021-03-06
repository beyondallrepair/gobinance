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

func TestClient_QueryOrderByClientID(t *testing.T) {
	t.Parallel()
	const (
		testSymbol        = "testSymbol"
		testClientOrderID = "someClientID"
	)
	call := func(ctx context.Context, uut *gobinance.Client, options []gobinance.QueryOrderOption) (gobinance.SpotOrder, error) {
		return uut.QueryOrderByClientID(ctx, testSymbol, testClientOrderID, options...)
	}

	cases := commonSpotQueryTestCases(func() url.Values {
		return map[string][]string{
			"symbol":            {testSymbol},
			"origClientOrderId": {testClientOrderID},
		}
	}, call)
	runQueryOrderTestCases(t, cases...)
}

func TestClient_QueryOrderByID(t *testing.T) {
	t.Parallel()
	const (
		testSymbol = "testSymbol"
		orderID    = 1234
	)
	call := func(ctx context.Context, uut *gobinance.Client, options []gobinance.QueryOrderOption) (gobinance.SpotOrder, error) {
		return uut.QueryOrderByID(ctx, testSymbol, orderID, options...)
	}

	cases := commonSpotQueryTestCases(func() url.Values {
		return map[string][]string{
			"symbol":  {testSymbol},
			"orderId": {fmt.Sprint(orderID)},
		}
	}, call)
	runQueryOrderTestCases(t, cases...)
}

func commonSpotQueryTestCases(expectedValues func() url.Values, call func(context.Context, *gobinance.Client, []gobinance.QueryOrderOption) (gobinance.SpotOrder, error)) []queryOrderTestCase {
	// note that some tests modify the expected url.Values in place.  Therefore, to avoid
	// races, the expectedValues function should return a newly constructed url.Values map
	// with the expected values in it.
	return []queryOrderTestCase{
		{
			name: "request values",
			setup: func(t *testing.T, mocks *clientMocks) {
				expectedValues := expectedValues()
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					if req.Method != http.MethodGet {
						t.Errorf("unexpected http method: expected %v but got %v", http.MethodGet, req.Method)
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
			name: "QueryOrderRecvWindow",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					expected := fmt.Sprint((testRecvWindow + time.Second).Milliseconds())
					if got := req.URL.Query().Get("recvWindow"); got != expected {
						t.Errorf("unexpected value for recvWindow. expected %v but got %v", expected, got)
					}
				}).Return(nil, fmt.Errorf("stop early"))
			},
			options: []gobinance.QueryOrderOption{
				gobinance.QueryOrderRecvWindow(testRecvWindow + time.Second),
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
					  "symbol": "LTCBTC",
					  "orderId": 1,
					  "orderListId": -1,
					  "clientOrderId": "myOrder1",
					  "price": "0.25",
					  "origQty": "1.25",
					  "executedQty": "3.25",
					  "cummulativeQuoteQty": "4.25",
					  "status": "NEW",
					  "timeInForce": "GTC",
					  "type": "LIMIT",
					  "side": "BUY",
					  "stopPrice": "5.25",
					  "icebergQty": "6.25",
					  "time": 1499827319559,
					  "updateTime": 1499827319560,
					  "isWorking": true,
					  "origQuoteOrderQty": "7.250000"
					}`)),
				}, nil)
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNil,
			expectedResult: gobinance.SpotOrder{
				Symbol:                "LTCBTC",
				OrderID:               1,
				OrderListID:           -1,
				ClientOrderID:         "myOrder1",
				Price:                 big.NewFloat(0.25),
				OriginalQty:           big.NewFloat(1.25),
				ExecutedQty:           big.NewFloat(3.25),
				CumulativeQuoteQty:    big.NewFloat(4.25),
				Status:                gobinance.OrderStatusNew,
				TimeInForce:           gobinance.TimeInForceGoodTilCanceled,
				Type:                  gobinance.OrderTypeLimit,
				Side:                  gobinance.OrderSideBuy,
				StopPrice:             big.NewFloat(5.25),
				IcebergQty:            big.NewFloat(6.25),
				Time:                  time.Date(2017, 07, 12, 02, 41, 59, int(559*time.Millisecond), time.UTC),
				UpdateTime:            time.Date(2017, 07, 12, 02, 41, 59, int(560*time.Millisecond), time.UTC),
				IsWorking:             true,
				OriginalQuoteOrderQty: big.NewFloat(7.25),
			},
		},
	}
}

type queryOrderTestCase struct {
	name           string
	setup          func(t *testing.T, mocks *clientMocks)
	ctx            context.Context
	options        []gobinance.QueryOrderOption
	call           func(ctx context.Context, uut *gobinance.Client, options []gobinance.QueryOrderOption) (gobinance.SpotOrder, error)
	errorCheck     errorCheck
	expectedResult gobinance.SpotOrder
}

func runQueryOrderTestCases(t *testing.T, testCases ...queryOrderTestCase) {
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
