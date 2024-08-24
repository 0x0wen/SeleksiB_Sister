const net = require('net');

class Sister {
  constructor() {
    this.routes = [];
    this.middlewares = [];
    this.server = net.createServer(socket => this.handleConnection(socket));
  }

  // Add a route
  addRoute(method, path, ...handlers) {
    this.routes.push({ method, path, handlers });
  }

  // Handle incoming connections
  handleConnection(socket) {
    socket.on('data', data => {
      const request = this.parseRequest(data.toString(), socket);
      this.executeMiddlewares(request, () => {
        this.handleRequest(request);
      });
    });

    socket.on('error', err => console.log('Socket Error:', err));
  }

  // Parse the incoming request
  parseRequest(data, socket) {

    // Split the request into headers and body
    const [headerPart, bodyPart] = data.split('\r\n\r\n');

    // Split headers into lines
    const headerLines = headerPart.split('\r\n');

    // Parse the request line (e.g., "GET /path HTTP/1.1")
    const [method, fullPath, protocol] = headerLines[0].split(' ');

    // Extract path and query parameters
    const { path, queryParams } = this.parseQueryParams(fullPath);

    // Parse headers into an object
    const headers = {};
    headerLines.slice(1).forEach(line => {
        const [key, value] = line.split(': ');
        headers[key.toLowerCase()] = value;
    });

    // Initialize body and file handling
    let body = null;
    let files = [];
    const contentType = headers['content-type'];

    if (contentType && contentType.startsWith('multipart/form-data')) {
        // Extract boundary from the Content-Type header
        const boundary = contentType.split('boundary=')[1];
        if (!boundary) {
            console.error('Boundary not found in Content-Type');
            return { method, path, queryParams, protocol, headers, body, files, socket };
        }

        // Split body by boundary
        const parts = bodyPart.split(`--${boundary}`);
        
        // Process each part
        parts.forEach(part => {
            // Skip empty parts or the final boundary marker
            if (!part || part.includes('--')) return;

            // Extract headers and body from the part
            const [partHeaders, partBody] = part.split('\r\n\r\n');
            const partHeaderLines = partHeaders.split('\r\n');

            // Extract Content-Disposition for file information
            const contentDisposition = partHeaderLines.find(line => line.startsWith('Content-Disposition:'));
            if (contentDisposition) {
                const [, disposition] = contentDisposition.split(': ');
                const match = /name="([^"]*)"; filename="([^"]*)"/.exec(disposition);
                if (match) {
                    const [_, name, filename] = match;
                    const contentType = partHeaderLines.find(line => line.startsWith('Content-Type:'))?.split(': ')[1];

                    if (filename) {
                        // This is a file part
                        files.push({ name, filename, contentType, data: partBody });
                    } else {
                        // This is a form field
                        body = { ...body, [name]: partBody };
                    }
                }
            }
        });
    } else {
        // Parse body based on other Content-Types
        if (bodyPart) {
            if (contentType === 'application/json') {
                try {
                    body = JSON.parse(bodyPart); // Parse JSON body
                } catch (error) {
                    console.error('Failed to parse JSON body:', error);
                }
            } else if (contentType === 'application/x-www-form-urlencoded') {
                body = {};
                bodyPart.split('&').forEach(part => {
                    const [key, value] = part.split('=');
                    body[decodeURIComponent(key)] = decodeURIComponent(value); // Parse URL-encoded body
                });
            } else if (contentType === 'text/plain') {
                body = bodyPart; // Parse plain text body
            }
        }
    }

    return {
        method,
        path,
        queryParams,
        protocol,
        headers,
        body,
        files,
        socket
    };
  }


  // Execute middlewares sequentially
  executeMiddlewares(req, finalHandler) {
    let index = -1;
    const next = () => {
      index++;
      if (index < this.middlewares.length) {
        this.middlewares[index](req, req.socket, next);
      } else {
        finalHandler();
      }
    };
    next();
  }

  // Handle the request by matching the route and executing the handlers
  handleRequest(req) {
    const route = this.routes.find(r => this.matchRoute(r, req));
    if (route) {
      req.params = this.extractParams(route.path, req.path);
      req.res = req.socket;
      const executeHandlers = (index = 0) => {
        if (index < route.handlers.length) {
          route.handlers[index](req, req.res, () => executeHandlers(index + 1));
        } else {
          req.socket.end();
        }
      };
      executeHandlers();
    } else {
      req.socket.write('HTTP/1.1 404 Not Found\r\n\r\n');
      req.socket.end();
    }
  }

  // Match the route with the request path
  matchRoute(route, req) {
    if (route.method !== req.method) return false;
    const routeSegments = route.path.split('/');
    const pathSegments = req.path.split('/');
    if (routeSegments.length !== pathSegments.length) return false;
    return routeSegments.every((seg, i) => seg.startsWith(':') || seg === pathSegments[i]);
  }

  // Extract dynamic parameters from the route
  extractParams(routePath, reqPath) {
    const params = {};
    const routeSegments = routePath.split('/');
    const pathSegments = reqPath.split('/');
    routeSegments.forEach((seg, i) => {
      if (seg.startsWith(':')) {
        const paramName = seg.slice(1);
        params[paramName] = pathSegments[i];
      }
    });
    return params;
  }

  // Parse query parameters from the URL
  parseQueryParams(url) {
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

  // Register GET route
  get(path, ...handlers) {
    this.addRoute('GET', path, ...handlers);
  }

  // Register POST route
  post(path, ...handlers) {
    this.addRoute('POST', path, ...handlers);
  }

  // Register PUT route
  put(path, ...handlers) {
    this.addRoute('PUT', path, ...handlers);
  }

  // Register DELETE route
  delete(path, ...handlers) {
    this.addRoute('DELETE', path, ...handlers);
  }

  // Register middleware
  use(middleware) {
    this.middlewares.push(middleware);
  }

  // Start the server
  listen(port) {
    this.server.listen(port, () => {
      console.log(`Server listening on port ${port}`);
    });
  }
}

module.exports = Sister;
