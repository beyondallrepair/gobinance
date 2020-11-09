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
	const (
		testBinanceApiKey = "test-binance-api-key"
		testUserAgent     = "test-user-agent"
		testBaseURL       = "https://example.com"
		defaultRecvWindow = 3 * time.Second
		currentTimeMillis = 1234567890123
		mockSignature     = "mock-signature"
	)

	var (
		mockCurrentTime = time.Unix(0, int64(currentTimeMillis*time.Millisecond))
		mockNow         = func() time.Time {
			return mockCurrentTime
		}
		testContext = context.WithValue(context.Background(), "test", "value")
	)

	errNil := func(t *testing.T, err error) bool {
		t.Helper()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return true
	}

	errNotNil := func(t *testing.T, err error) bool {
		t.Helper()
		if err == nil {
			t.Errorf("expectex an error but got nil")
		}
		return false
	}

	httpError := func(expectedStatus int, expectedErrCode int) func(*testing.T, error) bool {
		return func(t *testing.T, err error) bool {
			t.Helper()
			herr, ok := err.(interface {
				StatusCode() int
				ErrorCode() int
			})

			if !ok {
				t.Errorf("the error %#v did not implement the expected interface", err)
				return false
			}
			if gotStatus := herr.StatusCode(); gotStatus != expectedStatus {
				t.Errorf("unexpected status code.  expected %v but got %v", expectedStatus, gotStatus)
			}
			if gotErrCode := herr.ErrorCode(); gotErrCode != expectedErrCode {
				t.Errorf("unexpected error code.  expected %v but got %v", expectedErrCode, gotErrCode)
			}
			return false
		}
	}

	type mocks struct {
		*mock_gobinance.MockDoer
		*mock_gobinance.MockSigner
	}

	testCases := []struct {
		name           string
		testContext    context.Context
		setup          func(*testing.T, *mocks)
		errorCheck     func(t *testing.T, err error) bool
		expectedResult gobinance.AccountInformation
	}{
		{
			name:       "nil context",
			errorCheck: errNotNil,
			setup: func(t *testing.T, mocks *mocks) {
				params := fmt.Sprintf("recvWindow=%d&timestamp=%d", defaultRecvWindow.Milliseconds(), currentTimeMillis)
				mocks.MockSigner.EXPECT().Sign(params).Return(mockSignature)
			},
		},
		{
			name:        "binance error",
			testContext: testContext,
			setup: func(t *testing.T, mocks *mocks) {
				params := fmt.Sprintf("recvWindow=%d&timestamp=%d", defaultRecvWindow.Milliseconds(), currentTimeMillis)
				mocks.MockSigner.EXPECT().Sign(params).Return(mockSignature)
				signedParams := fmt.Sprintf("recvWindow=%d&signature=%s&timestamp=%d", defaultRecvWindow.Milliseconds(), mockSignature, currentTimeMillis)

				expectedRequest, _ := http.NewRequestWithContext(testContext, http.MethodGet, testBaseURL+"/api/v3/account?"+signedParams, nil)
				expectedRequest.Header.Set("User-Agent", testUserAgent)
				expectedRequest.Header.Set("X-MBX-APIKEY", testBinanceApiKey)

				mockResponse := &http.Response{
					StatusCode: 400,
					Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{ "code": %d, "msg": "doesn't matter" }`, -1234))),
				}
				mocks.MockDoer.EXPECT().Do(expectedRequest).Return(mockResponse, nil)
			},
			errorCheck: httpError(400, -1234),
		},
		{
			name:        "corrupt ok response",
			testContext: testContext,
			setup: func(t *testing.T, mocks *mocks) {
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
			setup: func(t *testing.T, mocks *mocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mockResponse := &http.Response{
					StatusCode: 400,
					Body:       ioutil.NopCloser(strings.NewReader(`not json`)),
				}
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(mockResponse, nil)
			},
			errorCheck: httpError(400, 0),
		},
		{
			name:        "request error",
			testContext: testContext,
			setup: func(t *testing.T, mocks *mocks) {
				mocks.MockSigner.EXPECT().Sign(gomock.Any()).Return(mockSignature)
				mocks.MockDoer.EXPECT().Do(gomock.Any()).Return(nil, fmt.Errorf("test error"))
			},
			errorCheck: errNotNil,
		},
		{
			name:        "ok",
			testContext: testContext,
			setup: func(t *testing.T, mocks *mocks) {
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
			mocks := &mocks{
				MockDoer:   doer,
				MockSigner: signer,
			}
			u, _ := url.Parse(testBaseURL)
			uut := &gobinance.Client{
				HTTPApiURL: u,
				UserAgent:  testUserAgent,
				APIKey:     testBinanceApiKey,
				RecvWindow: defaultRecvWindow,
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
