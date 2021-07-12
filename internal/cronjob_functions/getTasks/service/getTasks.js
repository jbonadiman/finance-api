const axios = require('axios');

exports.func = async () => {
    const apiUrl = process.env.API_URL;

    if (!apiUrl) {
        Error('undefined \'apiUrl\'!');
    }
    const auth = {
        username: process.env.USERNAME,
        password: process.env.PASSWORD,
    };

    if (!auth.username) {
        Error('undefined \'username\'!');
    }

    if (!auth.password) {
        Error('undefined \'password\'!');
    }

    const response = await axios.get(
        apiUrl, { auth });

    console.log(response.data);
}