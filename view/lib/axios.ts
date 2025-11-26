import axios from "axios";

export const api = axios.create({
  baseURL: "http://192.168.88.178:3010",
  headers: {
    "Content-Type": "application/json",
    Accept: "application/json",
  },
  withCredentials: true,
});
