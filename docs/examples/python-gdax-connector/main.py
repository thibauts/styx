import websocket
import json

STYX_URL = 'ws://localhost:8000/logs/gdax/records'
GDAX_URL = 'wss://ws-feed.gdax.com'

# Connect to GDAX Websocket feed and subscribe to the full BTC-USD order feed
gdax_ws = websocket.create_connection(GDAX_URL)

gdax_ws.send(json.dumps({
	'type': 'subscribe',
	'channels': ['full'],
	'product_ids': ['BTC-USD']
}));

# Connect to the "gdax" log on Styx and relay events
styx_ws = websocket.create_connection(STYX_URL, header=['X-HTTP-Method-Override: POST'])

for event in gdax_ws:
	styx_ws.send(event)
