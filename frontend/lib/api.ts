import axios, { AxiosInstance } from "axios";
import { getAccessToken } from "./utils";


let iaxios: AxiosInstance = axios.create({
    baseURL: "/z/api",
});

export const initHttpClient = () => {

    const token = getAccessToken();

    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Overhead": "Aaron Swartz",
    }

    if (token) {
        headers["Authorization"] = `TokenV1 ${token}`;
    }


    iaxios = axios.create({
        baseURL: "/z/api",
        headers,
    });


}



export const login = async (email: string, password: string) => {
    return iaxios.post<{ access_token: string }>("/login", {
        email,
        password,
    });
}

