
// const
import axios from 'axios';

const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImJxdWlubkBtYXRhZG9yYXBwLmNvbSIsImV4cCI6MjU0NjEzNTk4NCwidXNlcl9pZCI6ImU0OTE4OTgzLWY4YzEtNGE0YS1iODE4LWQ0YjMxMTQ5ZDZjNCIsInVzZXJuYW1lIjoiYmVuIn0.ZxMyCa03yitGrpLK3ZUZv490YAzERrVVnkVq-SoMIDU';
const baseURL = 'http://localhost:5000';

export default class http {

    static get(path) {
        return axios.get(`${baseURL}${path}`, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
    }

    static post(path, body) {
        return axios.post(`${baseURL}${path}`, body, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
    }

    static delete(path) {
        return axios.delete(`${baseURL}${path}`, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
    }

}
