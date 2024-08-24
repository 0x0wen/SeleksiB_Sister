const axios = require('axios');
const FormData = require('form-data');
const fs = require('fs');



// Function to test GET request
async function testGet() {
  try {
    const response = await axios.get('http://localhost:3000/nilai-akhir?q=123&n=anjaymabar&quill=anjay_hah123-312qwe');
    console.log('GET /nilai-akhir/123 Response:', response.data);
  } catch (error) {
    console.error('GET /nilai-akhir/123 Error:', error.response ? error.response.data : error.message);
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
    const response = await axios.post('http://localhost:3000/submit/a', {
      name: 'John Doe',
      car: 'Toyota',
      age: 30,
    }, {
      headers: {
        'Content-Type': 'application/json',
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
  await testGet();
  await testGetParam();
  await testPost();
  await testPostFile();
  await testPut();
  await testDelete();
}

runTests();
