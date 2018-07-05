package wumpus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"runtime"

	"github.com/gorilla/websocket"
)

const (
	// GatewayVersion indicates the version of Discord's Gateway to connect to
	GatewayVersion    = "6"
	largeThreshold    = 250
	stateDisconnected = iota
	stateConnecting
	stateConnected
	stateDisconnecting
)

// Shard is a shard connected to Discord's gateway
type Shard struct {
	token      string
	gateway    string
	ShardID    int
	shardTotal int
	state      int
	session    string
	ws         *websocket.Conn
	Done       chan struct{}
	Messages   chan []byte
}

// NewShard constructs a shard
func NewShard(token, gateway string, shardID, shardTotal int) Shard {
	return Shard{
		token:      token,
		gateway:    gateway,
		ShardID:    shardID,
		shardTotal: shardTotal,
		state:      stateDisconnected,
	}
}

// Start the shard, opening its socket
func (shard *Shard) Start() error {
	shard.Log("(inf) starting")
	if shard.state != stateDisconnected {
		shard.Log("(wrn) stopping existing state")
		err := shard.Stop()
		if err != nil {
			return err
		}
	}

	url := url.URL{
		Scheme: "wss",
		Host:   shard.gateway,
	}
	q := url.Query()
	q.Set("encoding", "json")
	q.Set("v", GatewayVersion)

	shard.state = stateConnecting
	shard.Done = make(chan struct{})
	shard.Messages = make(chan []byte, 16)

	c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil {
		shard.state = stateDisconnected
		return err
	}
	shard.ws = c
	c.SetCloseHandler(shard.disconnect)

	identify := Identify{
		Token: shard.token,
		Properties: Properties{
			OS:      runtime.GOOS,
			Browser: runtime.Version(),
			Device:  "adidis-wumpus 0",
		},
		Compress:       false,
		LargeThreshold: largeThreshold,
		Shard:          []int{shard.ShardID, shard.shardTotal},
	}
	err = shard.Send(OpIdentify, identify)
	if err != nil {
		shard.Stop()
		return err
	}

	shard.read()

	return nil
}

// Stop the shard, disconnecting any sockets
func (shard *Shard) Stop() error {
	shard.Log("(inf) stopping")
	// assign shard.ws to nil; even if there's an error closing it, gc will finish it off
	ws := shard.ws
	shard.ws = nil

	shard.state = stateDisconnecting
	close(shard.Done)
	close(shard.Messages)

	shard.state = stateDisconnected
	return ws.Close()
}

func (shard *Shard) disconnect(code int, text string) error {
	shard.Log("close", code, text)
	shard.Stop()
	return nil
}

func (shard *Shard) read() {
	go func() {
		for {
			_, buf, err := shard.ws.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			shard.Messages <- buf
		}
	}()
}

// Send a message to discord
func (shard *Shard) Send(op int, data interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	payload := Payload{
		Op:   op,
		Data: d,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	shard.Log("(sil) write op", op)
	err = shard.ws.WriteMessage(websocket.TextMessage, buf)
	return err
}

// Ready handles the shard's READY message
func (shard *Shard) Ready(payload []byte) {
	ready := new(Ready)
	err := json.Unmarshal(payload, &ready)
	if err != nil {
		log.Panicln("couldn't unmarshal ready", err)
	}
	shard.session = ready.SessionID
}

// Log a message from this shard
func (shard *Shard) Log(msg ...interface{}) {
	prefix := fmt.Sprintf("(S%d)", shard.ShardID)
	log.Println(prefix, msg)
}
