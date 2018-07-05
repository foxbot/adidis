package wumpus

import (
	"encoding/json"
)

const (
	opDispatch            = 0
	opHeartbeat           = 1
	opIdentify            = 2
	opStatusUpdate        = 3
	opVoiceStateUpdate    = 4
	opVoiceServerPing     = 5
	opResume              = 6
	opReconnect           = 7
	opRequestGuildMembers = 8
	opInvalidSession      = 9
	opHello               = 10
	opHeartbeatACK        = 11
)

type (
	// Payload structure for incoming or outgoing messages
	Payload struct {
		Op       int             `json:"op"`
		Data     json.RawMessage `json:"d"`
		Sequence int             `json:"s"`
		Type     string          `json:"t"`
	}

	// Hello is sent on socket open
	Hello struct {
		Heartbeat int `json:"heartbeat_interval"`
	}

	// Identify tells the socket who we are
	Identify struct {
		Token          string     `json:"token"`
		Properties     Properties `json:"properties"`
		Compress       bool       `json:"compress"`
		LargeThreshold int        `json:"large_threshold"`
		Shard          []int      `json:"shard"`
	}

	// Properties is embedded in Identify
	Properties struct {
		OS      string `json:"$os"`
		Browser string `json:"$browser"`
		Device  string `json:"$device"`
	}

	// Resume tries to resume
	Resume struct {
		Token    string `json:"token"`
		Session  string `json:"session_id"`
		Sequence int    `json:"seq"`
	}

	// Ready is sent when a session is created
	Ready struct {
		SessionID string `json:"session_id"`
	}
)
