package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Websocket struct {
	conn *websocket.Conn
}

func newWebSocket() *Websocket {
	// connect to websocket endpoint
	conn, _, err := websocket.DefaultDialer.Dial("wss://ws.kraken.com", http.Header{})
	if err != nil {
		log.Fatal("dial:", err)
	}
	//defer c.Close()
	return &Websocket{conn: conn}
}

func (s Websocket) Send() {
	defer s.conn.Close()
	var init_resp map[string]interface{}
	// get connection status
	err := s.conn.ReadJSON(&init_resp)
	if err != nil {
		log.Fatal(err)
	}
	if !(init_resp["event"] == "systemStatus" && init_resp["status"] == "online") {
		log.Fatal(init_resp)
	}

	// do subscription to pair
	payload := fmt.Sprintf(`{"event": "subscribe", "pair": ["XBT/USD","XBT/EUR"], "subscription": {"name": "ticker"}}`)
	s.conn.WriteMessage(1, []byte(payload))
	err = s.conn.ReadJSON(&init_resp)
	if err != nil {
		log.Fatal(err)
	} else if !(init_resp["event"] == "subscriptionStatus" && init_resp["pair"] == "XBT/USD" && init_resp["status"] == "subscribed") {
		log.Fatal(init_resp)
	}

	// listen to pair subscription
	resp := map[string]interface{}{}
	msg := []byte{}
	heartbeatResponse := []byte{123, 34, 101, 118, 101, 110, 116, 34, 58, 34, 104, 101, 97, 114, 116, 98, 101, 97, 116, 34, 125} // equals to {"event":"heartbeat"}

	fmt.Println(string(heartbeatResponse))
	// get initial state of the subscription with XBT/USD XBT/EUR pair
	for {
		_, msg, err = s.conn.ReadMessage()
		if err != nil {
			log.Fatal(err)
		} else if s.IsNotHeartbeat(heartbeatResponse, msg) { // healthcheck
			// not a heartbeatResponse message
			err := json.Unmarshal(msg, &resp)
			if err != nil {
				log.Fatal(err)
			}
			break
		}
	}

	resp1 := []interface{}{}
	// handle ticker message
	for {
		_, msg, err = s.conn.ReadMessage()
		if err != nil {
			log.Fatal(err)
		} else if s.IsNotHeartbeat(heartbeatResponse, msg) { // healthcheck
			// not a heartbeatResponse message
			err := json.Unmarshal(msg, &resp1)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", resp1)
		}
	}
}

func (s Websocket) IsNotHeartbeat(heartbeatResponse []byte, msg []byte) bool {
	return bytes.Compare(heartbeatResponse, msg) != 0
}

func main() {
	s := newWebSocket()
	s.Send()
}
