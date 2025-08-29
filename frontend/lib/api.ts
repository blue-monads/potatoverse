import axios, { AxiosInstance } from "axios";
import { getAccessToken } from "./utils";


const baseURL = window.location.origin + "/z/api";

let axiosInstance: AxiosInstance = axios.create({
    baseURL,
});

export const initAxios = () => {

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



export const login = async (email: string, password: string) => {
    return axiosInstance.post<{ access_token: string }>("/login", {
        email,
        password,
    });
}

