package gobinance

import (
	"encoding/json"
	"time"
)

type Balance struct {
	Asset  string
	Free   string
	Locked string
}

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
	// Balances is a map of asset to its balance
	Balances    map[string]Balance
	Permissions []string
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
