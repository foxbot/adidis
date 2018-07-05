package main

import (
	"flag"
	"log"

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

loop:
	for {
		select {
		case msg := <-shard.Messages:
			shard.Log("(sil) event", msg.Type)
			onDispatch(msg, &shard.ShardID)
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
