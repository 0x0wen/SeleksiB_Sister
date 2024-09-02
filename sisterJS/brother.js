const fs = require('fs');
const path = require('path');
const { stringifyJSON } = require('./helpers');

class Brother {
    constructor(socket) {
        this.socket = socket;
        this.headers = {
            'Content-Type': 'text/plain',
            'Connection': 'close',
        };
        this.statusCode = 200;
        this.statusMessage = 'OK';
        this.boundary = `--boundary-${Date.now()}`; // Unique boundary
    }

    setHeader(name, value) {
        this.headers[name] = value;
    }

    write(data) {
        this.socket.write(data);
    }

    end(data) {
        this.socket.end(data);
    }

    on(event, callback) {
        this.socket.on(event, callback);
    }

    writeHead(statusCode, statusMessage) {
        this.statusCode = statusCode;
        this.statusMessage = statusMessage;
    }

    send(data) {
        // Construct the response line
        const responseLine = `HTTP/1.1 ${this.statusCode} ${this.statusMessage}\r\n`;

        // Construct headers
        const headers = Object.entries(this.headers)
            .map(([key, value]) => `${key}: ${value}`)
            .join('\r\n');

        // Finalize the response
        const response = `${responseLine}${headers}\r\n\r\n`;

        // Send the response line and headers
        this.socket.write(response);

        // Send the response body
        this.socket.write(data, () => {
            this.socket.end();  // Close the connection
        });
    }

    sendMultipartResponse(jsonData, filePath) {
        this.setHeader('Content-Type', `multipart/form-data; boundary=${this.boundary}`);

        const startBoundary = `--${this.boundary}\r\n`;
        const endBoundary = `--${this.boundary}--\r\n`;

        // Construct the JSON part
        const jsonPart = `${startBoundary}Content-Disposition: form-data; name="jsonData"\r\n` +
                         `Content-Type: application/json\r\n\r\n` +
                         `${stringifyJSON(jsonData)}\r\n`;

        // Read the file
        const resolvedPath = path.resolve(filePath);

        fs.readFile(resolvedPath, (err, fileData) => {
            if (err) {
                this.writeHead(500, 'Internal Server Error');
                this.send('Error reading file');
                return;
            }

            const filePart = `${startBoundary}Content-Disposition: form-data; name="file"; filename="${path.basename(filePath)}"\r\n` +
                             `Content-Type: application/octet-stream\r\n\r\n`;

            // Send headers
            this.writeHead(200, 'OK');
            const responseLine = `HTTP/1.1 ${this.statusCode} ${this.statusMessage}\r\n`;

            const headers = Object.entries(this.headers)
                .map(([key, value]) => `${key}: ${value}`)
                .join('\r\n');

            this.socket.write(`${responseLine}${headers}\r\n\r\n`);

            // Send JSON part
            this.socket.write(jsonPart);

            // Send file part
            this.socket.write(filePart);
            this.socket.write(fileData);

            // Send end boundary
            this.socket.write(`\r\n${endBoundary}`, () => {
                this.socket.end(); // Close the connection
            });
        });
    }
}


module.exports = Brother;
