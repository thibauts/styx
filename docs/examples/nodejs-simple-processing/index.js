const Websocket = require('websocket-stream')
const Transform = require('stream').Transform

function transform() {

	const stats = {};

	return new Transform({
	    objectMode: true,
		transform: (data, _, done) => {
		
			event = JSON.parse(data);

			if('type' in event) {
				if(event.type in stats) {
					counter = stats[event.type]
					counter += 1
					stats[event.type] = counter
				} else {
					stats[event.type] = 1
				}
			}

			done(null, JSON.stringify(stats))
	    }
  });
}

const source = Websocket('ws://localhost:8000/logs/events/records?follow=true', {
	objectMode: true
})

const sink = Websocket('ws://localhost:8000/logs/stats/records', {
	objectMode: true,
	headers: {
		'X-HTTP-Method-Override': 'POST'
	}
})

source.pipe(transform()).pipe(sink)