package bitstamp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _cliId, _key, _secret string

var _url string = "https://www.bitstamp.net/api/v2"

type ErrorResult struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
	Code   string `json:"code"`
}

type AccountBalanceResult struct {
	Balance   map[string]float64
	Reserved  map[string]float64
	Available map[string]float64
	Fee       map[string]float64
}

func NewAccountBalanceResult() *AccountBalanceResult {
	return &AccountBalanceResult{
		Balance:   make(map[string]float64, 0),
		Available: make(map[string]float64, 0),
		Reserved:  make(map[string]float64, 0),
		Fee:       make(map[string]float64, 0)}
}

func (abr *AccountBalanceResult) UnmarshalJSON(data []byte) (err error) {
	var strMap map[string]string
	err = json.Unmarshal(data, &strMap)
	if err != nil {
		return
	}

	for key, value := range strMap {
		split := strings.Split(key, "_")
		if len(split) != 2 {
			return fmt.Errorf("could not identify key '%s'", key)
		}

		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}

		mapKey := split[0]
		mapID := split[1]

		switch mapID {
		case "balance":
			abr.Balance[mapKey] = floatValue
		case "reserved":
			abr.Reserved[mapKey] = floatValue
		case "available":
			abr.Available[mapKey] = floatValue
		case "fee":
			abr.Fee[mapKey] = floatValue
		default:
			return fmt.Errorf("Could not identify key postfix '%s'", mapID)
		}

	}

	return nil
}

type TickerResult struct {
	Last      float64 `json:"last,string"`
	High      float64 `json:"high,string"`
	Low       float64 `json:"low,string"`
	Vwap      float64 `json:"vwap,string"`
	Volume    float64 `json:"volume,string"`
	Bid       float64 `json:"bid,string"`
	Ask       float64 `json:"ask,string"`
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open,string"`
}

type BuyOrderResult struct {
	Id       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Price    float64 `json:"price,string"`
	Amount   float64 `json:"amount,string"`
}

type SellOrderResult struct {
	Id       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Price    float64 `json:"price,string"`
	Amount   float64 `json:"amount,string"`
}

type OrderBookResult struct {
	Timestamp string          `json:"timestamp"`
	Bids      []OrderBookItem `json:"bids"`
	Asks      []OrderBookItem `json:"asks"`
}

type OrderBookItem struct {
	Price  float64
	Amount float64
}

type OpenOrder struct {
	Id           int64   `json:"id,string"`
	DateTime     string  `json:"datetime"`
	Type         int     `json:"type,string"`
	Price        float64 `json:"price,string"`
	Amount       float64 `json:"amount,string"`
	CurrencyPair string  `json:"currency_pair"`
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
	if err != nil {
		return err
	}

	if bsEr, ok := e["error"]; ok {
		return fmt.Errorf("%v", bsEr)
	}

	// Check for status == error
	errResult := ErrorResult{}
	err = json.Unmarshal(body, &errResult)
	if err == nil {
		if errResult.Status == "error" {
			return fmt.Errorf("%#v", errResult)
		}
	}

	//parse the JSON response into the response object
	//log.Println(string(body))
	return json.Unmarshal(body, v)
}

// UnmarshalJSON takes a json array and converts it into an OrderBookItem.
func (o *OrderBookItem) UnmarshalJSON(data []byte) error {
	tmp_struct := struct {
		p string
		v string
	}{}

	err := json.Unmarshal(data, &[]interface{}{&tmp_struct.p, &tmp_struct.v})
	if err != nil {
		return err
	}

	if o.Price, err = strconv.ParseFloat(tmp_struct.p, 64); err != nil {
		return err
	}

	if o.Amount, err = strconv.ParseFloat(tmp_struct.v, 64); err != nil {
		return err
	}
	return nil
}

func AccountBalance() (*AccountBalanceResult, error) {
	balance := NewAccountBalanceResult()
	err := privateQuery("/balance/", url.Values{}, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func OrderBook(pair string) (*OrderBookResult, error) {
	orderBook := &OrderBookResult{}
	err := privateQuery("/order_book/"+pair+"/", url.Values{}, orderBook)
	if err != nil {
		return nil, err
	}
	return orderBook, nil
}

func Ticker(pair string) (*TickerResult, error) {
	ticker := &TickerResult{}
	err := privateQuery("/ticker/"+pair+"/", url.Values{}, ticker)
	if err != nil {
		return nil, err
	}
	return ticker, nil
}

func BuyLimitOrder(pair string, amount float64, price float64, amountPrecision, pricePrecision int) (*BuyOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', amountPrecision, 64))
	v.Add("price", strconv.FormatFloat(price, 'f', pricePrecision, 64))

	// make request
	result := &BuyOrderResult{}
	err := privateQuery("/buy/"+pair+"/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func BuyMarketOrder(pair string, amount float64) (*BuyOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', 8, 64))

	// make request
	result := &BuyOrderResult{}
	err := privateQuery("/buy/market/"+pair+"/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SellLimitOrder(pair string, amount float64, price float64, amountPrecision, pricePrecision int) (*SellOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', amountPrecision, 64))
	v.Add("price", strconv.FormatFloat(price, 'f', pricePrecision, 64))

	// make request
	result := &SellOrderResult{}
	err := privateQuery("/sell/"+pair+"/", v, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SellMarketOrder(pair string, amount float64) (*SellOrderResult, error) {
	// set params
	var v = url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', 8, 64))

	// make request
	result := &SellOrderResult{}
	err := privateQuery("/sell/market/"+pair+"/", v, result)
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

func CancelAllOrders() (*bool, error) {
	// make request

	log.Printf("CancelAllOrders() is currently untested!")

	var result *bool
	err := privateQuery("/cancel_all_orders/", url.Values{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func WithdrawLitecoin(address string, amount float64) (interface{}, error) {
	// make request

	log.Printf("WithdrawLitecoin() is currently untested!")

	var result map[string]interface{}

	v := url.Values{}
	v.Add("amount", strconv.FormatFloat(amount, 'f', 8, 64))
	v.Add("address", address)

	err := privateQuery("/ltc_withdrawal/", v, &result)

	if status, ok := result["status"]; ok {
		statusStr, ok := status.(string)
		if !ok {
			err = fmt.Errorf("Failed to deal with result '%+v'", result)
			return nil, err
		}

		if statusStr == "error" {
			err = fmt.Errorf("Got error: %+v", result)
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}
