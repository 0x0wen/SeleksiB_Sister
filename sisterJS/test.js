const axios = require('axios');
const FormData = require('form-data');
const fs = require('fs');

// Function to test GET request
async function testGet() {
  try {
    const response = await axios.get('http://localhost:3000/nilai-akhir?name=ahmad&age=23');
    console.log('GET /nilai-akhir/123 Response:', response.data);
  } catch (error) {
    console.log('uhuy')
    console.error('GET /nilai-akhir/123 Error:', error );
  }
}

async function testGetParam() {
  try {
    const response = await axios.get('http://localhost:3000/nilai-akhir/Ahmed');
    console.log('GET /nilai-akhir/Ahmed Response:', response.data);
  } catch (error) {
    console.error('GET /nilai-akhir/Ahmed Error:', error.response ? error.response.data : error.message);
  }
}
// Function to test POST request
async function testPost() {
  try {
      // Create a new FormData instance
      const form = new FormData();
      // Append the image file to the form data
      // fs.createReadStream('./image.png')
      form.append('image', fs.createReadStream('./image.png'));
      form.append('name', "Wick");
      form.append('car', "Mitsubishi");
      form.append('age', 23);
      const response = await axios.post('http://localhost:3000/submit/a', form, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    console.log('POST /user Response:', response.data);
  } catch (error) {
    console.error('POST /user Error:', error.response ? error.response.data : error.message);
  }
}

async function testPostFile() {
  try {
      // Create a new FormData instance
      const form = new FormData();
          // Append the image file to the form data
      form.append('image', fs.createReadStream('./image.png'));
      // Send the POST request with Axios
      const response = await axios.post('http://localhost:3000/submit-file/a', form, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    console.log('POST /user Response:', response.data);
  } catch (error) {
    console.error('POST /user Error:', error.response ? error.response.data : error.message);
  }
}

// Function to test PUT request
async function testPut() {
  try {
    const response = await axios.put('http://localhost:3000/update', {
      name: 'Jane Doe',
    }, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    console.log('PUT /user/123 Response:', response.data);
  } catch (error) {
    console.error('PUT /user/123 Error:', error.response ? error.response.data : error.message);
  }
}

// Function to test DELETE request
async function testDelete() {
  try {
    const response = await axios.delete('http://localhost:3000/delete/junaidi');
    console.log('DELETE /user/123 Response:', response.data);
  } catch (error) {
    console.error('DELETE /user/123 Error:', error.response ? error.response.data : error.message);
  }
}

// Run all tests
async function runTests() {
  const startTime = Date.now();
  let end;

  await Promise.all([testGet(),
  testGetParam(),
  testPost(),
  testPostFile(),
  testPut(),
  testDelete(),])

  endTime = Date.now();
  console.log(`All tests completed in ${endTime - startTime} milliseconds.`);
}

runTests();
