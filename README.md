# adidis

# what?
a distributed discord sharder.

this guy makes shards to connect to discord. distributes the messages across workers.

there's a [whitepaper](https://gist.github.com/DasWolke/c9d7dfe6a78445011162a12abd32091d)
about this design, it's pretty interesting and you should read it some time.

# why?

**scenario:** my discord bot has lots of big image commands. i use python because it has great
image manipulation libraries and a great discord library. but my bot uses so much memory and
the CPU is constantly pinned, so it runs really slow.

**problem:** you are getting bottlenecked by your own code, but because your bot is coupled so
tightly to discord, you can't really do anything about it. more shards = more complexity and more
points of failure.

**answer:** use a super fast language to talk to discord. keep a couple of these guys answering
dispatch events. drop the ones you don't care about, put the ones you do care about in a big
queue. cache the data you want in an in-memory database, get rid of the rest of it because it's
fat and wastes memory. spawn a million python workers to answer discord events, bring them up and
down as you choose, have basically zero downtime.

# who?

this design is owed to [zeyla](https://github.com/zeyla) and [wolke](https://github.com/daswolke)
who each have their own implementations of this system.

this one was written for fun, mostly.

# where?

no where, yet.

# when?

when someone asks.

# wlicense
Licensed under ISC. [LICENSE](LICENSE.md)