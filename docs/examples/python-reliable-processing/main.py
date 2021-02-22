import websocket
import requests
import json

STYX_SOURCE = 'gdax'
STYX_SINK = 'matches'

# Try to fetch the last processed position from the sink log: if not available, restart at 0.
res = requests.get('http://localhost:8000/logs/%s/records' % STYX_SINK, params={
	'position': -1,
	'whence': 'end'
})
if res.status_code == 200:
	last = res.json()
	position = last['position'] + 1
else:
	position = 0

print("restarting at position %d" % position)

# Setup source and sink websocket streams. Request source from restart position.
source = websocket.create_connection('ws://localhost:8000/logs/%s/records?position=%d&follow=true' % (STYX_SOURCE, position))
sink = websocket.create_connection('ws://localhost:8000/logs/%s/records' % STYX_SINK, header=[
	'X-HTTP-Method-Override: POST'
])

# Start the processing loop
for record in source:

	event = json.loads(record)

	# Filter and normalize events, keeping only matches.
	if event['type'] == 'match':

		# We store at which position the event was found in the source stream,
		# enabling us to restart from exactly this position in case of restart or crash.
		transformed = {
			'time': event['time'],
			'price': event['price'],
			'position': position,
		}

		sink.send(json.dumps(transformed))

	position += 1
