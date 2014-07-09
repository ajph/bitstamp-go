bitstamp-go
===========

A client implementation of the Bitstamp API, including websockets, in Golang.

Example Usage
-----

```go
package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/ajph/bitstamp-go"
)

func handleEvent(e *bitstamp.Event) {
	switch e.Event {
	// pusher stuff
	case "pusher:connection_established":
		log.Println("Connected")
	case "pusher_internal:subscription_succeeded":
		log.Println("Subscribed")
	case "pusher:pong":
		// ignore
	case "pusher:ping":
		Ws.Pong()

	// bitstamp
	case "trade":
		fmt.Printf("%#v\n", e.Data)

	// other
	default:
		log.Printf("Unknown event: %#v\n", e)
	}
}

func main() {

	// setup bitstamp api
	bitstamp.SetAuth("123456", "key", "secret")

	// get balance
	_, err := bitstamp.AccountBalance()
	if err != nil {
		fmt.Printf("Can't get balance using bitstamp API: %s\n", err)
		return
	}
	fmt.Println("\nAvailable Balances:")
	fmt.Printf("USD %f\n", balance.UsdAvailable)
	fmt.Printf("BTC %f\n", balance.BtcAvailable)	
	fmt.Printf("FEE %f\n\n", balance.Fee)

	// attempt to place a buy order
	order, err := bitstamp.BuyLimitOrder(0.5, 600.00)
	if err != nil {
		log.Printf("Error placing buy order: %s", err)
		return
	}
	
	// check order				
	var orderRes *bitstamp.OrderTransactionsResult									
	orderRes, err = bitstamp.OrderTransactions(order.Id)
	if err != nil {
		log.Printf("Error checking status of buy order #%d %s. Retrying...", order.Id, err)
		return
	}

	if orderRes.TotalBtcAmount != 0.5 {
		log.Printf("BUY order #%d unsuccessful", order.Id)
		return
	}

	// websocket read loop
	for {	
		// connect
		log.Println("Dialing...")
		var err error
		Ws, err = bitstamp.NewWebSocket(WS_TIMEOUT)
		if err != nil {
			log.Printf("Error connecting: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		Ws.Subscribe("live_trades")

		// read data
L:
		for {
			select {
			case ev := <-Ws.Stream:
				handleEvent(ev)

			case err := <-Ws.Errors:
				log.Printf("Socket error: %s, reconnecting...", err)
				Ws.Close()
				break L

			case <- time.After(10 * time.Second):
				Ws.Ping()

			}
		}
	}	

}
```

Todo
----
- Documentation
- Tests
