import axios from 'axios';

export default class HttpClient {

    constructor(host, token) {
        this.baseURL = `http://${host}`;
        this.token = token;
    }

    get(path) {
        return axios.get(`${this.baseURL}${path}`, {
            headers: {
                Authorization: `Bearer ${this.token}`
            }
        })
    }

    post(path, body) {
        return axios.post(`${this.baseURL}${path}`, body, {
            headers: {
                Authorization: `Bearer ${this.token}`
            }
        })
    }

    delete(path) {
        return axios.delete(`${this.baseURL}${path}`, {
            headers: {
                Authorization: `Bearer ${this.token}`
            }
        })
    }

}
