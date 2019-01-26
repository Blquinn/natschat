import axios from 'axios';

const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImJxdWlubkBtYXRhZG9yYXBwLmNvbSIsImV4cCI6MjU0NjEzNTk4NCwidXNlcl9pZCI6ImU0OTE4OTgzLWY4YzEtNGE0YS1iODE4LWQ0YjMxMTQ5ZDZjNCIsInVzZXJuYW1lIjoiYmVuIn0.ZxMyCa03yitGrpLK3ZUZv490YAzERrVVnkVq-SoMIDU';

export default class HttpClient {

    constructor(host) {
        this.baseURL = `http://${host}`;
    }

    get(path) {
        return axios.get(`${this.baseURL}${path}`, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
    }

    post(path, body) {
        return axios.post(`${this.baseURL}${path}`, body, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
    }

    delete(path) {
        return axios.delete(`${this.baseURL}${path}`, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
    }

}
