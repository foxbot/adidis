package main

import (
	"flag"
	"log"

	"github.com/foxbot/adidis/wumpus"
	"github.com/mediocregopher/radix.v2/pool"
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
	redisPool  *pool.Pool
)

var greenEvents = []string{
	"READY",
	"MESSAGE_CREATE",
	"GUILD_CREATE",
	"GUILD_UPDATE",
	"GUILD_DELETE",
	"CHANNEL_CREATE",
	"CHANNEL_UPDATE",
	"CHANNEL_DELETE",
}

func init() {
	flag.StringVar(&token, "token", "", "Discord bot token")
	flag.IntVar(&shardID, "shardid", 0, "")
	flag.IntVar(&shardTotal, "shardtotal", 0, "")
}

func main() {
	flag.Parse()
	println("adidis")

	r, err := pool.New("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		panic(err)
	}
	redisPool = r

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
			onDispatch(msg)
		case <-shard.Done:
			break loop
		}
	}

	return shard.Stop()
}

func onDispatch(event *wumpus.Event) {
	greenlit := isGreenlit(event.Type)

	if !greenlit {
		return
	}
	forwardEvent(event)
}

func forwardEvent(event *wumpus.Event) error {
	redis, err := redisPool.Get()
	defer redisPool.Put(redis)

	if err != nil {
		return err
	}

	resp := redis.Cmd("RPUSH", "exchange:events", event.Data)
	return resp.Err
}

func isGreenlit(event string) bool {
	for _, ev := range greenEvents {
		if event == ev {
			return true
		}
	}
	return false
}
