# import websocket

# STYX_SOURCE = 'gdax'
# position = 0

# source = websocket.create_connection('ws://localhost:8000/logs/%s/records?position=%d' % (STYX_SOURCE, position))

# for record in source:
# 	print(record)

# import aiohttp
# import asyncio

# STYX_SOURCE = 'gdax'
# position = 0

# async def main():

# 	async with aiohttp.ClientSession() as session:

# 		ws = await session.ws_connect('ws://localhost:8000/logs/%s/records?position=%d&follow=true' % (STYX_SOURCE, position))

# 		while True:

# 			msg = await ws.receive()

# 			# print(msg)

# 			if msg.type == aiohttp.WSMsgType.CLOSED:
# 				break
# 			elif msg.type == aiohttp.WSMsgType.ERROR:
# 				break

# 			record = msg.data
# 			print(record)

# loop = asyncio.get_event_loop()
# loop.run_until_complete(main())

import asyncio
import websockets

STYX_SOURCE = 'gdax'
position = 0

async def main():

	uri = 'ws://localhost:8000/logs/%s/records?position=%d&follow=false' % (STYX_SOURCE, position)

	async with websockets.connect(uri) as source:

		async for message in source:
			print(message)



loop = asyncio.get_event_loop()
loop.run_until_complete(main())
