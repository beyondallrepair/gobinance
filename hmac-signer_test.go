package gobinance_test

import (
	"github.com/beyondallrepair/gobinance"
	"testing"
)

func TestHMACSigner_Sign(t *testing.T) {
	t.Parallel()
	const (
		// from the binance docs
		binanceExampleKey = "NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j"
	)
	testCases := []struct {
		name           string
		key            string
		input          string
		expectedOutput string
	}{
		{
			name: "binance example 1",
			key: binanceExampleKey,
			input: "symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559",
			expectedOutput:  "c8db56825ae71d6d79447849e617115f4a920fa2acdcab2b053c4b2838bd6b71",
		},
		{
			name: "binance example 2",
			key: binanceExampleKey,
			input: "symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559",
			expectedOutput: "c8db56825ae71d6d79447849e617115f4a920fa2acdcab2b053c4b2838bd6b71",
		},
		{
			name: "binance example 3",
			key: binanceExampleKey,
			input: "symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTCquantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559",
			expectedOutput: "0fd168b8ddb4876a0358a8d14d0c9f3da0e9b20c5d52b2a00fcf7d1c602f9a77",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			uut := &gobinance.HMACSigner{Secret: tc.key}
			got := uut.Sign(tc.input)
			if got != tc.expectedOutput {
				t.Errorf("unexpected signature.  expected\n\t%v\ngot\n\t%v", tc.expectedOutput, got)
			}
		})
	}
}
