package wumpus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"time"

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
	token         string
	gateway       string
	ShardID       int
	shardTotal    int
	state         int
	session       string
	sequence      int
	ws            *websocket.Conn
	heartbeatDone chan struct{}
	Done          chan struct{}
	Messages      chan *Payload
}

// NewShard constructs a shard
func NewShard(token, gateway string, shardID, shardTotal int) Shard {
	return Shard{
		token:         token,
		gateway:       gateway,
		ShardID:       shardID,
		shardTotal:    shardTotal,
		state:         stateDisconnected,
		heartbeatDone: make(chan struct{}, 1),
		Done:          make(chan struct{}, 1),
		Messages:      make(chan *Payload, 32),
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

	c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil {
		shard.state = stateDisconnected
		return err
	}
	shard.ws = c
	c.SetCloseHandler(shard.disconnect)

	shard.read()

	if shard.session != "" {
		err = shard.resume()
		if err != nil {
			shard.Stop()
			return err
		}
		return nil
	}

	err = shard.identify()
	if err != nil {
		shard.Stop()
		return err
	}

	return nil
}

func (shard *Shard) identify() error {
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
	err := shard.Send(opIdentify, identify)
	return err
}

func (shard *Shard) resume() error {
	resume := Resume{
		Token:    shard.token,
		Session:  shard.session,
		Sequence: shard.sequence,
	}
	err := shard.Send(opResume, resume)
	return err
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
	shard.Log("(inf) connection lost", code, text)

	if code == 1000 || code == 1001 {
		shard.sequence = 0
		shard.session = ""
	}

	shard.heartbeatDone <- struct{}{}

	shard.Stop()
	return nil
}

func (shard *Shard) read() {
	go func() {
		for {
			_, buf, err := shard.ws.ReadMessage()
			if err != nil {
				shard.Log("(err) socket read failed", err)
				return
			}
			payload := new(Payload)
			err = json.Unmarshal(buf, &payload)
			if err != nil {
				shard.Log("(wrn) json unmarshal failed, dropping packet", err)
				continue
			}
			switch payload.Op {
			case opDispatch:
				shard.sequence = payload.Sequence
				shard.Messages <- payload

				if payload.Type == "READY" {
					shard.handleReady(payload.Data)
				}
			case opHello:
				shard.handleHello(payload.Data)
			case opInvalidSession:
				shard.identify()
			}
		}
	}()
}

func (shard *Shard) handleHello(data []byte) {
	hello := new(Hello)
	err := json.Unmarshal(data, &hello)
	if err != nil {
		log.Panicln("couldn't unmarshal hello", err)
	}

	shard.heartbeatDone <- struct{}{}
	close(shard.heartbeatDone)
	shard.heartbeatDone = make(chan struct{}, 1)

	ticker := time.NewTicker(time.Duration(hello.Heartbeat) * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				shard.Send(opHeartbeat, struct{}{})
			case <-shard.heartbeatDone:
				ticker.Stop()
				break
			}
		}
	}()
}
func (shard *Shard) handleReady(data []byte) {
	ready := new(Ready)
	err := json.Unmarshal(data, &ready)
	if err != nil {
		log.Panicln("couldn't unmarshal ready", err)
	}
	shard.session = ready.SessionID
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

// Log a message from this shard
func (shard *Shard) Log(msg ...interface{}) {
	prefix := fmt.Sprintf("(S%d)", shard.ShardID)
	log.Println(prefix, msg)
}
