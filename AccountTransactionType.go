package bitstamp

import (
	"strconv"
	"strings"
	"time"
)

const (
	UserDeposit UserTransactionType = iota
	UserWithdrawal
	UserMarketTrade
	UserSubAccountTransfer
)

type UserTransactionType int8

func (t *UserTransactionType) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	i, err := strconv.ParseInt(s, 10, 8)
	*t = UserTransactionType(i)
	return err
}

func (t UserTransactionType) String() string {
	switch t {
	case UserDeposit:
		return "Deposit"
	case UserWithdrawal:
		return "Withdrawal"
	case UserMarketTrade:
		return "MarketTrade"
	case UserSubAccountTransfer:
		return "SubAccountTransfer"
	default:
		return ""
	}
}

type accountTransactionsResult struct {
	DateTime Time                `json:"datetime"`
	Id       int64               `json:"id"`
	Type     UserTransactionType `json:"type"`
	Usd      Float               `json:"usd"`
	Eur      Float               `json:"eur"`
	Btc      Float               `json:"btc"`
	Xrp      Float               `json:"xrp"`
	Ltc      Float               `json:"ltc"`
	Eth      Float               `json:"eth"`
	BtcUsd   Float               `json:"btc_usd"`
	UsdBtc   Float               `json:"usd_btc"`
	Fee      Float               `json:"fee"`
	OrderId  int64               `json:"order_id"`
}

type AccountTransactionResult struct {
	DateTime time.Time           `json:"datetime"`
	Id       int64               `json:"id"`
	Type     UserTransactionType `json:"type"`
	Usd      float64             `json:"usd"`
	Eur      float64             `json:"eur"`
	Btc      float64             `json:"btc"`
	Xrp      float64             `json:"xrp"`
	Ltc      float64             `json:"ltc"`
	Eth      float64             `json:"eth"`
	BtcUsd   float64             `json:"btc_usd"`
	UsdBtc   float64             `json:"usd_btc"`
	Fee      float64             `json:"fee"`
	OrderId  int64               `json:"order_id"`
}
