from discord.http import HTTPClient, Route

import asyncio
import json
import redis
import sys

# Set this to blaze URL
Route.BASE = "http://localhost:15001"

class Client:
    def __init__(self):
        self.handlers = {}
        self.redis = redis.StrictRedis()

    async def call(self, type, data):
        if type in self.handlers:
            for h in self.handlers[type]:
                func = h(data)
                if asyncio.iscoroutinefunction(h):
                    await func

    def event(self, type):
        def registerhandler(handler):
            if type in self.handlers:
                self.handlers[type].append(handler)
            else:
                self.handlers[type] = [handler]
            return handler
        return registerhandler

    async def run(self):
        while True:
            _, raw = self.redis.blpop('exchange:events')
            ev = json.loads(raw)
            await self.call(ev['t'], ev['d'])

client = Client()

# Cache Management
@client.event('READY')
def ready(data):
    client.redis.set('discord:me', json.dumps(data['user']))

@client.event('CHANNEL_CREATE')
def channel_create(data):
    client.redis.set('discord:channels:%s' % data['id'], json.dumps(data))

@client.event('CHANNEL_UPDATE')
def channel_update(data):
    client.redis.set('discord:channels:%s' % data['id'], json.dumps(data))

@client.event('CHANNEL_DELETE')
def channel_delete(data):
    client.redis.delete('discord:channels:%s' % data['id'])

@client.event('GUILD_CREATE')
def guild_create(data):
    client.redis.set('discord:guilds:%s' % data['id'], json.dumps(data))
    for channel in data['channels']:
        client.redis.set('discord:channels:%s' % channel['id'], json.dumps(channel))

# Commands
me = None
prefix = ''

http = HTTPClient()

@client.event('MESSAGE_CREATE')
def message(data):
    channel = json.loads(client.redis.get('discord:channels:%s' % data['channel_id']))
    print("%s in %s: %s" % (data['author']['username'], channel['name'], data['content']))

@client.event('MESSAGE_CREATE')
async def command(data):
    global me
    global prefix

    if me is None:
        me = json.loads(client.redis.get('discord:me'))
        prefix = '<@%s> ' % me['id']
    content = data['content']

    if not content.startswith(prefix):
        return

    command = content[len(prefix):]
    print('trying to find a command %s' % command)

    if command == 'ping':
        print('sending a big pong')
        await http.send_message(data['channel_id'], 'pong')


loop = asyncio.get_event_loop()
loop.run_until_complete(client.run())
loop.close()
