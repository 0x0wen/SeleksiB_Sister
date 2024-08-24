const net = require('net');
const routes = {};

const server = net.createServer(socket => {
  socket.on('data', data => {
    const request = data.toString();
    handleRequest(socket, request);
  });

  socket.on('error', err => console.log('Socket Error:', err));
});

function handleRequest(socket, request) {
  const [requestLine, ...headers] = request.split('\r\n');
  const [method, fullPath] = requestLine.split(' ');
  
  const { path, queryParams } = parseQueryParams(fullPath);
  const routeHandler = routes[`${method}:${path}`];
  if (routeHandler) {
    const response = routeHandler(queryParams);
    socket.write(response);
  } else {
    socket.write('HTTP/1.1 404 Not Found\r\n\r\n');
  }
  socket.end();
}

function parseQueryParams(url) {
  const [path, query] = url.split('?');
  const queryParams = {};
  if (query) {
    query.split('&').forEach(param => {
      const [key, value] = param.split('=');
      queryParams[key] = value;
    });
  }
  return { path, queryParams };
}

function createRoute(method, path, handler) {
  routes[`${method}:${path}`] = handler;
}

function get(path, handler) {
    createRoute('GET', path, handler);
  }
  
  function post(path, handler) {
    createRoute('POST', path, handler);
  }
  
  function put(path, handler) {
    createRoute('PUT', path, handler);
  }
  
  function del(path, handler) {  // 'delete' is a reserved word in JS
    createRoute('DELETE', path, handler);
  }
function startServer(port) {
  server.listen(port, () => {
    console.log(`Server listening on port ${port}`);
  });
}

module.exports = { createRoute, startServer };
