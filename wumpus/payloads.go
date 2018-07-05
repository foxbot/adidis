package wumpus

import (
	"encoding/json"
)

const (
	// OpDispatch is for dispatch events
	OpDispatch = 0
	// OpHeartbeat is a heartbeat
	OpHeartbeat = 1
	// OpIdentify is an outgoing identify
	OpIdentify = 2
	// OpStatusUpdate sets the bot status
	OpStatusUpdate     = 3
	OpVoiceStateUpdate = 4
	OpVoiceServerPing  = 5
	// OpResume asks for the session to be resumed
	OpResume              = 6
	OpReconnect           = 7
	OpRequestGuildMembers = 8
	OpInvalidSession      = 9
	OpHello               = 10
	OpHeartbeatACK        = 11
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
