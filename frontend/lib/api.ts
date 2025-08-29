import axios from "axios";
import { getAccessToken } from "./utils";


let axiosInstance = null;

const initAxios = () => {

    const baseURL = window.location.origin + "/z/api";
    const token = getAccessToken();

    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Overhead": "Aaron Swartz",
    }

    if (token) {
        headers["Authorization"] = `TokenV1 ${token}`;
    }
    

    axiosInstance = axios.create({
        baseURL,
        headers,
    });
    

}


