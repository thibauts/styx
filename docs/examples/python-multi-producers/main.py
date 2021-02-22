import sys
import random
import json
from datetime import datetime
import asyncio
import aiohttp

STYX_URL = 'http://localhost:8000/logs/measures/records'

async def producer(producer_id):

	# Delay producer startup by a random value between 0 and 1 second.
	await asyncio.sleep(random.random())

	async with aiohttp.ClientSession() as session:
		while True:

			# Generate a random temperature data point
			temperature = random.random() * 30
			timestamp = datetime.now().isoformat()

			event = json.dumps({
				'timestamp': timestamp,
				'sensor': producer_id,
				'temperature': temperature
			})

			# Push it to the "measures" log
			await session.post(STYX_URL, data=event)

			# print('pushed event', event)

			# Wait one second before generating another
			await asyncio.sleep(1)


async def main():

	if len(sys.argv) != 2:
		print('Usage: python3 main.py <PRODUCER COUNT>')
		return

	count = int(sys.argv[1])

	await asyncio.gather(*[producer('producer-%d' % (i+1)) for i in range(count)])
	

loop = asyncio.get_event_loop()
loop.run_until_complete(main())
