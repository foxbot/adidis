import json
import redis

class Client:
    def __init__(self):
        self.handlers = {}
        self.redis = redis.StrictRedis()

    def call(self, type, data):
        if type in self.handlers:
            for h in self.handlers[type]:
                h(data)

    def event(self, type):
        def registerhandler(handler):
            if type in self.handlers:
                self.handlers[type].append(handler)
            else:
                self.handlers[type] = [handler]
            return handler
        return registerhandler

    def run(self):
        while True:
            _, raw = self.redis.blpop('exchange:events')
            ev = json.loads(raw)
            self.call(ev['t'], ev['d'])

client = Client()

@client.event('MESSAGE_CREATE')
def message(data):
    print("%s in %s: %s" % (data['author']['username'], data['channel_id'], data['content']))

client.run()
