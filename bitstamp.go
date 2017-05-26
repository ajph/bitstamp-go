package bitstamp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _cliId, _key, _secret string

var _url string = "https://www.bitstamp.net/api/v2"

type AccountBalanceResult struct {
	UsdBalance   float64 `json:"usd_balance,string"`
	BtcBalance   float64 `json:"btc_balance,string"`
	EurBalance   float64 `json:"eur_balance,string"`
	XrpBalance   float64 `json:"xrp_balance,string"`
	UsdReserved  float64 `json:"usd_reserved,string"`
	BtcReserved  float64 `json:"btc_reserved,string"`
	EurReserved  float64 `json:"eur_reserved,string"`
	XrpReserved  float64 `json:"xrp_reserved,string"`
	UsdAvailable float64 `json:"usd_available,string"`
	BtcAvailable float64 `json:"btc_available,string"`
	EurAvailable float64 `json:"eur_available,string"`
	XrpAvailable float64 `json:"xrp_available,string"`
	BtcUsdFee    float64 `json:"btcusd_fee,string"`
	BtcEurFee    float64 `json:"btceur_fee,string"`
	EurUsdFee    float64 `json:"eurusd_fee,string"`
	XrpUsdFee    float64 `json:"xrpusd_fee,string"`
	XrpEurFee    float64 `json:"xrpeur_fee,string"`
	XrpBtcFee    float64 `json:"xrpbtc_fee,string"`
}

type TickerResult struct {
	Last         float64 `json:"last,string"`
	High	     float64 `json:"high,string"`
	Low          float64 `json:"low,string"`
	Vwap         float64 `json:"vwap,string"`
	Volume	     float64 `json:"volume,string"`
	Bid          float64 `json:"bid,string"`
	Ask          float64 `json:"ask,string"`
	Timestamp    string  `json:"timestamp"`
	Open         float64 `json:"open,string"`
}

type BuyLimitOrderResult struct {
	Id       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Price    float64 `json:"price,string"`
	Amount   float64 `json:"amount,string"`
}

type SellLimitOrderResult struct {
	Id       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Price    float64 `json:"price,string"`
	Amount   float64 `json:"amount,string"`
}

type UserTransactionResult struct {
	Id       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Usd      float64 `json:"usd,string"`
	Btc      float64 `json:"btc,string"`
	Fee      float64 `json:"fee,string"`
	OrderId  int64   `json:"order_id,string"`
}

type OpenOrder struct {
	Id       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Price    float64 `json:"price,string"`
	Amount   float64 `json:"amount,string"`
	CurrencyPair string `json:"currency_pair"`
}

type OrderTransactionsResult struct {
	TotalFee       float64
	TotalUsdAmount float64
	TotalBtcAmount float64
	UsdPerBtc      float64
}

func SetAuth(clientId, key, secret string) {
	_cliId = clientId
	_key = key
	_secret = secret
}

// privateQuery submits an http.Request with key, sig & nonce
func privateQuery(path string, values url.Values, v interface{}) error {
	// parse the bitstamp URL
	endpoint, err := url.Parse(_url)
	if err != nil {
		return err
	}

	// set the endpoint for this request
	endpoint.Path += path

	// add required key, signature & nonce to values
	nonce := strconv.FormatInt(time.Now().UnixNano(), 10)
	mac := hmac.New(sha256.New, []byte(_secret))
	mac.Write([]byte(nonce + _cliId + _key))
	values.Set("key", _key)
	values.Set("signature", strings.ToUpper(hex.EncodeToString(mac.Sum(nil))))
	values.Set("nonce", nonce)

	// encode the url.Values in the body
	reqBody := strings.NewReader(values.Encode())

	// create the request
	//log.Println(endpoint.String(), values)
	req, err := http.NewRequest("POST", endpoint.String(), reqBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// submit the http request
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// if no result interface, return
	if v == nil {
		return nil
	}

	// read the body of the http message into a byte array
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	// is this an error?
	if len(body) == 0 {
		return fmt.Errorf("Response body 0 length")
	}
	e := make(map[string]interface{})
	err = json.Unmarshal(body, &e)
	if bsEr, ok := e["error"]; ok {
		return fmt.Errorf("%v", bsEr)
	}

	//parse the JSON response into the response object
	//log.Println(string(body))
	return json.Unmarshal(body, v)
}

func AccountBalance() (*AccountBalanceResult, error) {
	balance := &AccountBalanceResult{}
	err := privateQuery("/balance/", url.Values{}, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func Ticker(pair string) (*TickerResult, error) {
	ticker := &TickerResult{}
	err := privateQuery("/ticker/" + pair + "/", url.Values{}, ticker)
	if err != nil {
		return nil, err
	}
	return ticker, nil
}

func BuyLimitOrder(pair string, amount float64, price float64, amountPrecision, pricePrecision int) (*BuyLimitOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', amountPrecision, 64))
	v.Add("price", strconv.FormatFloat(price, 'f', pricePrecision, 64))

	// make request
	result := &BuyLimitOrderResult{}
	err := privateQuery("/buy/" + pair + "/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func BuyMarketOrder(pair string, amount float64) (*BuyLimitOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', 8, 64))

	// make request
	result := &BuyLimitOrderResult{}
	err := privateQuery("/buy/market/" + pair + "/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SellLimitOrder(pair string, amount float64, price float64, amountPrecision, pricePrecision int) (*SellLimitOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', amountPrecision, 64))
	v.Add("price", strconv.FormatFloat(price, 'f', pricePrecision, 64))

	// make request
	result := &SellLimitOrderResult{}
	err := privateQuery("/sell/" + pair + "/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SellMarketOrder(pair string, amount float64) (*SellLimitOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', 8, 64))

	// make request
	result := &SellLimitOrderResult{}
	err := privateQuery("/sell/market/" + pair + "/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CancelOrder(orderId int64) {
	// set params
	var v = url.Values{}
	v.Add("id", strconv.FormatInt(orderId, 10))

	// make request
	privateQuery("/cancel_order/", v, nil)
}

func OpenOrders() (*[]OpenOrder, error) {
	// make request
	result := &[]OpenOrder{}
	err := privateQuery("/open_orders/all/", url.Values{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func UserTransactions(offset int64, limit int64, sort string) ([]UserTransactionResult, error) {
	// set params
	var v = url.Values{}
	v.Add("offset", strconv.FormatInt(offset, 10))
	v.Add("limit", strconv.FormatInt(limit, 10))
	v.Add("sort", sort)

	// make request
	result := &[]UserTransactionResult{}
	err := privateQuery("/user_transactions/", v, result)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// checks the past 100 transactions and sums results for a specified orderid
func OrderTransactions(orderId int64) (*OrderTransactionsResult, error) {
	ut, err := UserTransactions(0, 500, "desc")
	if err != nil {
		return nil, err
	}
	ot := &OrderTransactionsResult{}
	for i := 0; i < len(ut); i++ {
		if ut[i].OrderId == orderId {
			ot.TotalFee += math.Abs(ut[i].Fee)
			ot.TotalUsdAmount += math.Abs(ut[i].Usd)
			ot.TotalBtcAmount += math.Abs(ut[i].Btc)
		}
	}
	ot.UsdPerBtc = ot.TotalUsdAmount / ot.TotalBtcAmount
	return ot, nil
}
