const getAllData = async () => {
    console.log("Getting all data...");
    const data = [{name:"John Doe", car:"Toyota", age:30}, {name:"Jane Doe", car:"Honda", age:25}];
    return data
}

const getDataByName = async (name) => {
    console.log("Getting data by name...");
    const data = {name:name, car:"Toyota", age:10};
    return data;
}

const createData = async (name,car,age,file) => {
    console.log("Creating data...");
    const data = {name:name, car:car, age:age,file:file};
    return data;
};

  const updateData = async (name,car,age) => {
    console.log("Updating data...");
    const data = {name:name, car:car, age:age};
    return data;
};

  const deleteData = async (name) => {
    console.log("Deleting data...");
    const data = {name:name, car:"Jaguar", age:12};
    return data;
};

  const duplicateData = async (name,car,age) => {
    console.log("Duplicating data...");
    const data = {name:name, car:car, age:age};
    return data;
};

const handleFile = async (file) =>{
    console.log("Handling file...");
    return file;
}

  module.exports = { getAllData, getDataByName, createData, updateData, deleteData, duplicateData,handleFile };