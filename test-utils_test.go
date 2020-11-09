package gobinance_test

import (
	mock_gobinance "github.com/beyondallrepair/gobinance/mocks"
	"github.com/google/go-cmp/cmp"
	"math/big"
	"testing"
	"time"
)

// constants commonly used in tests
const (
	// currentTimeMillis is a mock of the current time in millisecnds.
	// This is equivalent to Fri Feb 13 2009 23:31:30
	currentTimeMillis = 1234567890123
	testBinanceApiKey = "test-binance-api-key"
	testUserAgent     = "test-user-agent"
	testBaseURL       = "https://example.com"
	testRecvWindow    = 3 * time.Second
	mockSignature     = "mock-signature"
)

// bigFloatComparer is used by google compare's function in order to allow
// checking for equality of big Floats
var bigFloatComparer = cmp.Comparer(func(a, b *big.Float) bool {
	return a.Cmp(b) == 0
})

func mustParseBigFloat(t *testing.T, str string) *big.Float {
	v, _, err := new(big.Float).Parse(str, 10)
	if err != nil {
		t.Fatalf("unable to parse big float %v: %v", str, err)
	}
	return v
}

// mockNow returns the mocked current time.
// when converted into milliseconds since the unix epoch, this must be equal to currentTimeMillis
func mockNow() time.Time {
	return time.Date(2009, 02, 13, 23, 31, 30, int(123*time.Millisecond), time.UTC)
}

type clientMocks struct {
	*mock_gobinance.MockDoer
	*mock_gobinance.MockSigner
}

// errorCheck performs some tests on an error result and Errors those tests if
// the value is not as expected.  It returns true if the test should continue
// after the check, false otherwise.  This is usually used to avoid checking
// other return values in the event that an error is expected, as usually the
// other results of an errored state is irrelevant.
type errorCheck func(t *testing.T, err error) bool

// errNil causes a test error if `err` is not nil
func errNil(t *testing.T, err error) bool {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	return true
}

// errNotNil causes a test error if `err` is nil
func errNotNil(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error but got nil")
	}
	return false
}

// isHttpError returns a check that causes a test error if the error does not have a
// StatusCode() and ErrorCode() function that returns the specified values.
func isHttpError(expectedStatus int, expectedErrCode int) func(*testing.T, error) bool {
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
