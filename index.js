const http = require('http');

const server = http.createServer((req, res) => {
    console.log("Server will take 10 seconds for processing")
    setTimeout(() => {
        res.writeHead(200, { 'Content-Type': 'text/plain' });
        res.end('Delayed response after 3 seconds!\n');
      }, 10000);    
});

const PORT = 8000;

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}/`);
});
