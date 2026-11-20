const http = require("node:http");

const server = http.createServer((_request, response) => {
  response.writeHead(200, { "content-type": "text/plain" });
  response.end("This Node.js runtime is shipped in the container image.\n");
});

server.listen(8081, "0.0.0.0");
