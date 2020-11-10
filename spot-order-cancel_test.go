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

func TestClient_CancelOrderByClientOrderID(t *testing.T) {
	t.Parallel()
	const (
		testSymbol        = "testSymbol"
		testClientOrderID = "someClientID"
	)
	call := func(ctx context.Context, uut *gobinance.Client, options []gobinance.CancelSpotOrderOption) (gobinance.CancelSpotOrderResult, error) {
		return uut.CancelOrderByClientOrderID(ctx, testSymbol, testClientOrderID, options...)
	}

	cases := commonSpotCancelTestCases(func() url.Values {
		return map[string][]string{
			"symbol":            {testSymbol},
			"origClientOrderId": {testClientOrderID},
		}
	}, call)
	runCancelOrderTestCases(t, cases...)
}

func TestClient_CancelOrderByOrderID(t *testing.T) {
	t.Parallel()
	const (
		testSymbol  = "testSymbol"
		testOrderID = 12345
	)
	call := func(ctx context.Context, uut *gobinance.Client, options []gobinance.CancelSpotOrderOption) (gobinance.CancelSpotOrderResult, error) {
		return uut.CancelOrderByOrderID(ctx, testSymbol, testOrderID, options...)
	}

	cases := commonSpotCancelTestCases(func() url.Values {
		return map[string][]string{
			"symbol":  {testSymbol},
			"orderId": {fmt.Sprint(testOrderID)},
		}
	}, call)
	runCancelOrderTestCases(t, cases...)
}

type cancelOrderTestCase struct {
	name           string
	setup          func(t *testing.T, mocks *clientMocks)
	ctx            context.Context
	options        []gobinance.CancelSpotOrderOption
	call           func(ctx context.Context, uut *gobinance.Client, options []gobinance.CancelSpotOrderOption) (gobinance.CancelSpotOrderResult, error)
	errorCheck     errorCheck
	expectedResult gobinance.CancelSpotOrderResult
}

func commonSpotCancelTestCases(expectedValues func() url.Values, call func(context.Context, *gobinance.Client, []gobinance.CancelSpotOrderOption) (gobinance.CancelSpotOrderResult, error)) []cancelOrderTestCase {
	// note that some tests modify the expected url.Values in place.  Therefore, to avoid
	// races, the expectedValues function should return a newly constructed url.Values map
	// with the expected values in it.
	return []cancelOrderTestCase{
		{
			name: "request values",
			setup: func(t *testing.T, mocks *clientMocks) {
				expectedValues := expectedValues()
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					if req.Method != http.MethodDelete {
						t.Errorf("unexpected http method: expected %v but got %v", http.MethodDelete, req.Method)
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
			name: "cancelSpotClientOrderID",
			setup: func(t *testing.T, mocks *clientMocks) {
				expectedValues := expectedValues()
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					if val := req.URL.Query().Get("newClientOrderId"); val != "testOrderID" {
						t.Errorf("unexpected newClientOrderId. expected %v but got %v", "testOrderID", val)
					}
					// confirm we've not accidentally clobbered any of the other IDs
					expectedOrigClientOrderId := expectedValues.Get("origClientOrderId")
					if got := req.URL.Query().Get("origClientOrderId"); got != expectedOrigClientOrderId {
						t.Errorf("unexpected origClientOrderId. expected %v but got %v", expectedOrigClientOrderId, got)
					}
					expectedOrderId := expectedValues.Get("orderId")
					if got := req.URL.Query().Get("orderId"); got != expectedOrderId {
						t.Errorf("unexpected orderId. expected %v but got %v", expectedOrderId, got)
					}
				}).Return(nil, fmt.Errorf("stop early"))
			},
			options: []gobinance.CancelSpotOrderOption{
				gobinance.CancelSpotClientOrderID("testOrderID"),
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNotNil,
		},
		{
			name: "CancelSpotOrderRecvWindow",
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Do(func(req *http.Request) {
					expected := fmt.Sprint((testRecvWindow + time.Second).Milliseconds())
					if got := req.URL.Query().Get("recvWindow"); got != expected {
						t.Errorf("unexpected value for recvWindow. expected %v but got %v", expected, got)
					}
				}).Return(nil, fmt.Errorf("stop early"))
			},
			options: []gobinance.CancelSpotOrderOption{
				gobinance.CancelSpotOrderRecvWindow(testRecvWindow + time.Second),
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
					  "origClientOrderId": "myOrder1",
					  "orderId": 4,
					  "orderListId": -1,
					  "clientOrderId": "cancelMyOrder1",
					  "price": "2.25000000",
					  "origQty": "1.25000000",
					  "executedQty": "3.25000000",
					  "cummulativeQuoteQty": "4.25000000",
					  "status": "CANCELED",
					  "timeInForce": "GTC",
					  "type": "LIMIT",
					  "side": "BUY"
					}`)),
				}, nil)
			},
			ctx:        context.Background(),
			call:       call,
			errorCheck: errNil,
			expectedResult: gobinance.CancelSpotOrderResult{
				Symbol:                "LTCBTC",
				OrderID:               4,
				OrderListID:           -1,
				ClientOrderID:         "cancelMyOrder1",
				OriginalClientOrderID: "myOrder1",
				Price:                 big.NewFloat(2.25),
				OriginalQty:           big.NewFloat(1.25),
				ExecutedQty:           big.NewFloat(3.25),
				CumulativeQuoteQty:    big.NewFloat(4.25),
				Status:                gobinance.OrderStatusCanceled,
				TimeInForce:           gobinance.TimeInForceGoodTilCanceled,
				Type:                  gobinance.OrderTypeLimit,
				Side:                  gobinance.OrderSideBuy,
			},
		},
	}
}

func runCancelOrderTestCases(t *testing.T, testCases ...cancelOrderTestCase) {
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
