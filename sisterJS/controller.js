const Service = require('./service.js');


const getAllIdentity = async (req, res) => {
    try{
        console.log("ni paramnya abangkuh:", JSON.stringify(req.queryParams));
        const data = await Service.getAllData();
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
        res.end(JSON.stringify(data));
    }catch(error){
        console.log(error); 
        res.end(JSON.stringify(error));
    }
}

const getIdentityByName = async (req, res) => {
    try{
        const { name } = req.params;
        const data = await Service.getDataByName(name);
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
        res.end(JSON.stringify(data));
    }catch(error){
        console.log(error); 
        res.end(JSON.stringify(error));
    }
}

const createIdentity = async (req, res) => {
    try{
        const { name, car, age } = req.body;
        const data = await Service.createData(name, car, age);
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
        res.end(JSON.stringify(data));
    }catch(error){
        console.log(error); 
        res.end(JSON.stringify(error));
    }
}

const createIdentityFile = async (req, res) => {
    try{
        const file  = req.files[0];
        const data = await Service.handleFile(file);
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
        res.end(JSON.stringify(data));
    }catch(error){
        console.log(error); 
        res.end(JSON.stringify(error));
    }
}

const updateIdentity = async (req, res) => {
    try{
        const { ...reqData } = req.body;
        const data = await Service.updateData(reqData);
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
        res.end(JSON.stringify(data));
    }catch(error){
        console.log(error);
        res.end(JSON.stringify(error));
    }
}

const deleteIdentity = async (req, res) => {
    try{
        const { name } = req.params;
        const data = await Service.deleteData(name);
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
        res.end(JSON.stringify(data));
    }catch(error){
        console.log(error); 
        res.end(JSON.stringify(error));
    }
}

const duplicateIdentity = async (req, res) => {
    try{
        const { name, car, age } = req.body;
        const data = await Service.duplicateData(name, car, age);
        res.write('HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n');
    }catch(error){
        console.log(error); // Log the error for debugging
        res.end(JSON.stringify(error));
    }
}

module.exports = { getAllIdentity, getIdentityByName, createIdentity,createIdentityFile, updateIdentity, deleteIdentity, duplicateIdentity };