const net = require('net');
const path_lib = require('path');
const fs = require('fs');
const Brother = require('./brother');

class Sister {
  constructor() {
    this.routes = [];
    this.middlewares = [];
    this.server = net.createServer(socket => {
      const response = new Brother(socket); 
      this.handleConnection(response)
    });
  }

  // Add a route
  addRoute(method, path, ...handlers) {
    this.routes.push({ method, path, handlers });
  }

  // Handle incoming connections
  handleConnection(response) {
    
    let accumulatedData = '';
    response.on('data', chunk => {
      // if the data sent is too large, it will be split into multiple chunks
      // so we need to accumulate the data until it's complete, then we can parse it
      accumulatedData += chunk.toString('binary');
      console.log('accumulatedData', accumulatedData);
      // Check if the end of headers and start of body is present
      if (accumulatedData.includes('\r\n\r\n')) {
        // Get Content-Length from headers if available
        const [headerPart] = accumulatedData.split('\r\n\r\n');
        const contentLengthMatch = headerPart.match(/Content-Length: (\d+)/);
        if (contentLengthMatch) {
            const contentLength = parseInt(contentLengthMatch[1], 10);
            // Check if we have received enough data (headers + body)
            console.log('data length', accumulatedData.length);
            console.log('content length', contentLength);
            console.log('header length', headerPart.length);
            if (accumulatedData.length >= contentLength + headerPart.length + 4) {
                console.log('Received complete request:', accumulatedData);
                const request = this.parseRequest(accumulatedData, response);
                console.log('ahmad',request)
                this.executeMiddlewares(request, () => {
                    this.handleRequest(request);
                });

                accumulatedData = ''; // Clear buffer for next request
            }
        } else {
            // If no Content-Length, assume headers and body are received
            console.log('Received request:', accumulatedData);

            const request = this.parseRequest(accumulatedData, response);
            this.executeMiddlewares(request, () => {
                this.handleRequest(request);
            });

            accumulatedData = ''; // Clear buffer for next request
        }
    }
      // console.log('Received:', chunk.toString());
      // const request = this.parseRequest(chunk.toString(), socket);
      // this.executeMiddlewares(request, () => {
      //   this.handleRequest(request);
      // });
    });

    response.socket.on('error', err => {
      console.log('Socket Error:', err);
      accumulatedData = ''; // Clear buffer when connection ends
    });
  }

  // Parse the incoming request
  parseRequest(data, response) {
    console.log('mentahan', data);
    // Split the request into headers and body
    const [headerPart, ...rest] = data.split('\r\n\r\n');
    const bodyPart = rest.join('\r\n\r\n');

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
    let body = {};
    let files = [];
    const contentType = headers['content-type'];

    if (contentType && contentType.startsWith('multipart/form-data')) {
        // Extract boundary from the Content-Type header
        const boundary = contentType.split('boundary=')[1];
        if (!boundary) {
            console.error('Boundary not found in Content-Type');
            return { method, path, queryParams, protocol, headers, body, files, response };
        }

        // Split body by boundary
        const parts = bodyPart.split(`--${boundary}`);
        console.log("ini parts",parts);
        // Process each part
        parts.forEach(part => {
            // Skip empty parts or the final boundary marker
            if (!part) return;
            console.log('part', part);
            // Extract headers and body from the part
            const [partHeaders, partBody] = part.split('\r\n\r\n');
            console.log('partHeaders', partHeaders);
            console.log('partBody', partBody);
            const partHeaderLines = partHeaders.split('\r\n');

            // Extract Content-Disposition for file information
            const contentDisposition = partHeaderLines.find(line => line.startsWith('Content-Disposition:'));
            if (contentDisposition) {
                const nameMatch = contentDisposition.match(/name="([^"]+)"/);
                const filenameMatch = contentDisposition.match(/filename="([^"]+)"/);
                if (nameMatch) {
                    const name = nameMatch[1];

                    if (filenameMatch) {
                        // Handle file upload
                        const filename = filenameMatch[1];
                        const fileData = partBody.slice(0, -2); // Remove trailing \r\n
                       
                        const filePath = path_lib.join(__dirname, 'uploads', filename);

                        // Save the file to disk
                        fs.writeFileSync(filePath, fileData,'binary');

                        body[name] = {
                            filename: filename,
                            path: filePath,
                            contentType: headerLines.find(line => line.startsWith('Content-Type')).split(': ')[1]
                        };
                    } else {
                        // Handle regular form field
                        body[name] = partBody.slice(0, -2); // Remove trailing \r\n
                    }
                }
            }
        });
    } else if (contentType === 'application/x-www-form-urlencoded') {
        // Parse URL-encoded body
        body = {};
        bodyPart.split('&').forEach(part => {
            const [key, value] = part.split('=');
            body[decodeURIComponent(key)] = decodeURIComponent(value);
        });
    } else if (contentType === 'application/json') {
        try {
            body = JSON.parse(bodyPart);
        } catch (error) {
            console.error('Failed to parse JSON body:', error);
        }
    } else if (contentType === 'text/plain') {
        body = bodyPart;
    }

    return {
        method,
        path,
        queryParams,
        protocol,
        headers,
        body,
        files,
        response
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
      const executeHandlers = (index = 0) => {
        if (index < route.handlers.length) {
          route.handlers[index](req, req.response, () => executeHandlers(index + 1));
        } else {
          req.response.end();
        }
      };
      executeHandlers();
    } else {
      req.response.write('HTTP/1.1 404 Not Found\r\n\r\n');
      req.response.end();
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
