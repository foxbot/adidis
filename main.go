package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/foxbot/adidis/wumpus"
)

const (
	eventDrop = iota
	eventForward
	eventCache
)

var (
	token      string
	shardID    int
	shardTotal int
)

var eventMap = map[string]int{
	"READY":          eventForward,
	"MESSAGE_CREATE": eventForward,
}

func init() {
	flag.StringVar(&token, "token", "", "Discord bot token")
	flag.IntVar(&shardID, "shardid", 0, "")
	flag.IntVar(&shardTotal, "shardtotal", 0, "")
}

func main() {
	flag.Parse()
	println("adidis")

	// TODO: make lots and lots of shards
	shard := wumpus.NewShard(token, "gateway.discord.gg", 0, 1)
	log.Fatalln(runShard(shard))
}

func runShard(shard wumpus.Shard) error {
	err := shard.Start()
	if err != nil {
		return err
	}

	// assume that discord will always send a HELLO (this is probably really bad)
	msg := <-shard.Messages
	payload := new(wumpus.Payload)
	err = json.Unmarshal(msg, &payload)
	if err != nil {
		return err
	}
	interval := 0
	if payload.Op != wumpus.OpHello {
		shard.Log("(wrn) first frame was not a hello, falling back to default interval")
		interval = 45000
	} else {
		hello := new(wumpus.Hello)
		err = json.Unmarshal(payload.Data, &hello)
		if err != nil {
			return err
		}
		interval = hello.Heartbeat
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

loop:
	for {
		select {
		case msg := <-shard.Messages:
			payload := new(wumpus.Payload)
			err = json.Unmarshal(msg, &payload)
			if err != nil {
				shard.Log("(wrn) bad json, dropping packet", err)
				continue
			}

			switch payload.Op {
			case wumpus.OpDispatch:
				shard.Log("(sil) event", payload.Type)
				onDispatch(payload, &shard.ShardID)

				if payload.Type == "READY" {
					shard.Ready(payload.Data)
				}
			}
		case <-ticker.C:
			err = shard.Send(wumpus.OpHeartbeat, nil)
			if err != nil {
				shard.Log("(wrn) failed to heartbeat", err)
			}
		case <-shard.Done:
			break loop
		}
	}

	return shard.Stop()
}

func onDispatch(payload *wumpus.Payload, shardID *int) {
	action := getAction(payload.Type)

	switch action {
	case eventDrop:
		return
	case eventForward:
		// TODO: write to redis
	case eventCache:
		// TODO: cache in redis
	}
}

func getAction(event string) int {
	action, ok := eventMap[event]
	if ok {
		return action
	}
	return eventDrop
}
