package gobinance_test

import (
	"context"
	"fmt"
	"github.com/beyondallrepair/gobinance"
	mock_gobinance "github.com/beyondallrepair/gobinance/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHttpClient_AccountInformation(t *testing.T) {
	t.Parallel()

	type testContextKey string
	var (
		mockCurrentTime = time.Unix(0, int64(currentTimeMillis*time.Millisecond))
		mockNow         = func() time.Time {
			return mockCurrentTime
		}
		testContext = context.WithValue(context.Background(), testContextKey("test"), "value")
	)

	testCases := []struct {
		name           string
		testContext    context.Context
		setup          func(*testing.T, *clientMocks)
		errorCheck     errorCheck
		expectedResult gobinance.AccountInformation
	}{
		{
			name:       "nil context",
			errorCheck: errNotNil,
			setup: func(t *testing.T, mocks *clientMocks) {
				params := fmt.Sprintf("recvWindow=%d&timestamp=%d", testRecvWindow.Milliseconds(), currentTimeMillis)
				mocks.MockSigner.EXPECT().Sign(params).Return(mockSignature)
			},
		},
		{
			name:        "binance error",
			testContext: testContext,
			setup: func(t *testing.T, mocks *clientMocks) {
				params := fmt.Sprintf("recvWindow=%d&timestamp=%d", testRecvWindow.Milliseconds(), currentTimeMillis)
				mocks.MockSigner.EXPECT().Sign(params).Return(mockSignature)
				signedParams := fmt.Sprintf("recvWindow=%d&signature=%s&timestamp=%d", testRecvWindow.Milliseconds(), mockSignature, currentTimeMillis)

				expectedRequest, _ := http.NewRequestWithContext(testContext, http.MethodGet, testBaseURL+"/api/v3/account?"+signedParams, nil)
				expectedRequest.Header.Set("User-Agent", testUserAgent)
				expectedRequest.Header.Set("X-MBX-APIKEY", testBinanceApiKey)

				mockResponse := &http.Response{
					StatusCode: 400,
					Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{ "code": %d, "msg": "doesn't matter" }`, -1234))),
				}
				mocks.MockDoer.EXPECT().Do(expectedRequest).Return(mockResponse, nil)
			},
			errorCheck: isHttpError(400, -1234),
		},
		{
			name:        "corrupt ok response",
			testContext: testContext,
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mockResponse := &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(strings.NewReader(`not json`)),
				}
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(mockResponse, nil)
			},
			errorCheck: errNotNil,
		},
		{
			name:        "corrupt error response",
			testContext: testContext,
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mockResponse := &http.Response{
					StatusCode: 400,
					Body:       ioutil.NopCloser(strings.NewReader(`not json`)),
				}
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(mockResponse, nil)
			},
			errorCheck: isHttpError(400, 0),
		},
		{
			name:        "request error",
			testContext: testContext,
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(nil, fmt.Errorf("test error"))
			},
			errorCheck: errNotNil,
		},
		{
			name:        "ok",
			testContext: testContext,
			setup: func(t *testing.T, mocks *clientMocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode: 200,
					Body: ioutil.NopCloser(strings.NewReader(`
						{
						  "makerCommission": 1,
						  "takerCommission": 2,
						  "buyerCommission": 3,
						  "sellerCommission": 4,
						  "canTrade": true,
						  "canWithdraw": true,
						  "canDeposit": true,
						  "updateTime": 123456789321,
						  "accountType": "SPOT",
						  "balances": [
							{
							  "asset": "BTC",
							  "free": "4723846.89208129",
							  "locked": "1.00000000"
							},
							{
							  "asset": "LTC",
							  "free": "4763368.68006011",
							  "locked": "2.00000000"
							}
						  ],
							"permissions": [
							"SPOT"
						  ]
						}
					`)),
				}, nil)
			},
			errorCheck: errNil,
			expectedResult: gobinance.AccountInformation{
				MakerCommission:  1,
				TakerCommission:  2,
				BuyerCommission:  3,
				SellerCommission: 4,
				CanTrade:         true,
				CanWithdraw:      true,
				CanDeposit:       true,
				UpdateTime:       time.Unix(0, int64(123456789321*time.Millisecond)),
				AccountType:      "SPOT",
				Balances: map[string]gobinance.Balance{
					"BTC": {
						Asset:  "BTC",
						Free:   mustParseBigFloat(t, "4723846.89208129"),
						Locked: mustParseBigFloat(t, "1.00000000"),
					},
					"LTC": {
						Asset:  "LTC",
						Free:   mustParseBigFloat(t, "4763368.68006011"),
						Locked: mustParseBigFloat(t, "2.00000000"),
					},
				},
				Permissions: []string{"SPOT"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			doer := mock_gobinance.NewMockDoer(ctrl)
			signer := mock_gobinance.NewMockSigner(ctrl)
			mocks := &clientMocks{
				MockDoer:   doer,
				MockSigner: signer,
			}
			u, _ := url.Parse(testBaseURL)
			uut := &gobinance.Client{
				HTTPApiURL: u,
				UserAgent:  testUserAgent,
				APIKey:     testBinanceApiKey,
				RecvWindow: testRecvWindow,
				Signer:     signer,
				Doer:       doer,
				Now:        mockNow,
			}

			tc.setup(t, mocks)
			got, err := uut.AccountInformation(tc.testContext)
			if checkResult := tc.errorCheck(t, err); !checkResult {
				return
			}

			if diff := cmp.Diff(tc.expectedResult, got, bigFloatComparer); diff != "" {
				t.Errorf("unexpected result: %v", diff)
			}
		})
	}
}
