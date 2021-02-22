const http = require('http');
const url = require('url');

const server = http.createServer((req, res) => {

	const now = new Date()
	const query = url.parse(req.url, true).query;

	const event = {
		time: now.toISOString(),
		...query
	}

	data = JSON.stringify(event)

	const options = {
		hostname: 'localhost',
		port: 8000,
		path: '/logs/events/records',
		method: 'POST',
		headers: {
			'Content-Length': data.length,
		}
	};

	const styxReq = http.request(options, styxRes => {
		if(styxRes.statusCode != 200) {
			console.error(`couldn't push to styx (status=${styxRes.statusCode})`)
		}
	});

	styxReq.on('error', error => {
		console.error(`failed pushing event (error=${error})`)
	});

	styxReq.write(data)
	styxReq.end()

 	res.end();
});

server.listen(9000);