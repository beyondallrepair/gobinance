package gobinance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// AccountInformation fetches data about the account associated with the API Key provided to
// the client, or an error in the event of connection issues or a non-200 response from binance.
func (c *Client) AccountInformation(ctx context.Context) (AccountInformation, error) {
	req, err := c.buildSignedRequest(ctx, http.MethodGet, "/api/v3/account", nil)
	if err != nil {
		return AccountInformation{}, fmt.Errorf("error building request: %w", err)
	}
	var out AccountInformation
	if err := performRequest(c.Doer, req, &out); err != nil {
		return out, err
	}
	return out, nil
}

// Balance represents the amount of funds associated with a particular asset in this binance
// account
type Balance struct {
	// Asset is the symbol of the asset in question
	Asset string
	// Free is the amount of that asset that is currently available for trading
	Free *big.Float
	// Locked is the amount of that asset that is not available, due to being associated
	// with other open orders.
	Locked *big.Float
}

// AccountInformation is the response of an AccountInformation request from binance
type AccountInformation struct {
	MakerCommission  int64
	TakerCommission  int64
	BuyerCommission  int64
	SellerCommission int64
	CanTrade         bool
	CanWithdraw      bool
	CanDeposit       bool
	UpdateTime       time.Time
	AccountType      string
	Balances         map[string]Balance
	Permissions      []string
}

func (t *AccountInformation) UnmarshalJSON(bs []byte) error {
	var tmp struct {
		MakerCommission  int64           `json:"makerCommission"`
		TakerCommission  int64           `json:"takerCommission"`
		BuyerCommission  int64           `json:"buyerCommission"`
		SellerCommission int64           `json:"sellerCommission"`
		CanTrade         bool            `json:"canTrade"`
		CanWithdraw      bool            `json:"canWithdraw"`
		CanDeposit       bool            `json:"canDeposit"`
		UpdateTime       millisTimestamp `json:"updateTime"`
		AccountType      string          `json:"accountType"`
		Balances         []Balance       `json:"balances"`
		Permissions      []string        `json:"permissions"`
	}

	if err := json.Unmarshal(bs, &tmp); err != nil {
		return err
	}

	balanceMap := make(map[string]Balance)
	for _, b := range tmp.Balances {
		balanceMap[b.Asset] = b
	}
	*t = AccountInformation{
		MakerCommission:  tmp.MakerCommission,
		TakerCommission:  tmp.TakerCommission,
		BuyerCommission:  tmp.BuyerCommission,
		SellerCommission: tmp.SellerCommission,
		CanTrade:         tmp.CanTrade,
		CanWithdraw:      tmp.CanWithdraw,
		CanDeposit:       tmp.CanDeposit,
		UpdateTime:       time.Time(tmp.UpdateTime),
		AccountType:      tmp.AccountType,
		Balances:         balanceMap,
		Permissions:      tmp.Permissions,
	}
	return nil
}
