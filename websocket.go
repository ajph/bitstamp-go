package bitstamp

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
)

var _socketurl string = "wss://ws.pusherapp.com/app/de504dc5763aeef9ff52?protocol=7&client=js&version=2.1.6&flash=false"

type WebSocket struct {
	ws     *websocket.Conn
	quit   chan bool
	Stream chan *Event
	Errors chan error
}

type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func (s *WebSocket) Close() {
	s.quit <- true
}

func (s *WebSocket) Subscribe(channel string) {
	a := &Event{
		Event: "pusher:subscribe",
		Data: map[string]interface{}{
			"channel": channel,
		},
	}
	s.ws.WriteJSON(a)
}

func (s *WebSocket) SendTextMessage(message []byte) {
	s.ws.WriteMessage(websocket.TextMessage, message)
}

func (s *WebSocket) Ping() {
	a := &Event{
		Event: "pusher:ping",
	}
	s.ws.WriteJSON(a)
}

func (s *WebSocket) Pong() {
	a := &Event{
		Event: "pusher:pong",
	}
	s.ws.WriteJSON(a)
}

func NewWebSocket(t time.Duration) (*WebSocket, error) {
	var err error
	s := &WebSocket{
		quit:   make(chan bool, 1),
		Stream: make(chan *Event),
		Errors: make(chan error),
	}

	// set up websocket
	s.ws, _, err = websocket.DefaultDialer.Dial(_socketurl, nil)
	if err != nil {
		return nil, fmt.Errorf("error dialing websocket: %s", err)
	}

	go func() {
		defer s.ws.Close()
		for {
			runtime.Gosched()
			s.ws.SetReadDeadline(time.Now().Add(t))
			select {
			case <-s.quit:
				return
			default:
				var message []byte
				var err error
				_, message, err = s.ws.ReadMessage()
				if err != nil {
					s.Errors <- err
					continue
				}
				e := &Event{}
				err = json.Unmarshal(message, e)
				if err != nil {
					s.Errors <- err
					continue
				}
				s.Stream <- e
			}
		}
	}()

	return s, nil
}
