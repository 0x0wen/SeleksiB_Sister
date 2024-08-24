const Sister = require('./class');
const app = new Sister();
const Controller = require('./controller');


// Middleware example
function middlewareFunction1(req, res, next) {
  console.log('Middleware 1 executed');
  next();
}

function middlewareFunction2(req, res, next) {
  console.log('Middleware 2 executed');
  next();
}

// Define a GET route with middleware and a controller
app.get('/nilai-akhir',  Controller.getAllIdentity);

// Define a GET route with middleware and a controller
app.get('/nilai-akhir/:name',  Controller.getIdentityByName);

app.post('/submit/:id', middlewareFunction1, middlewareFunction2, Controller.createIdentity);
app.post('/submit-file/:id', middlewareFunction1, middlewareFunction2, Controller.createIdentityFile);

app.put('/update', Controller.updateIdentity);

app.delete('/delete/:name', Controller.deleteIdentity);

// Start the server on port 3000
app.listen(3000);
