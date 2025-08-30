import axios, { AxiosInstance } from "axios";
import { getLoginData } from "./utils";


let iaxios: AxiosInstance = axios.create({
    baseURL: "/z/api",
});

export const initHttpClient = () => {

    const data = getLoginData();

    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Overhead": "Aaron Swartz",
    }

    if (data?.accessToken) {
        headers["Authorization"] = `TokenV1 ${data.accessToken}`;
    }


    iaxios = axios.create({
        baseURL: "/z/api",
        headers,
    });


}


export const login = async (username: string, password: string) => {
    return iaxios.post<{ access_token: string, user_info: User }>("/core/auth/login", {
        username,
        password,
    });
}

export interface User {
    id: number;
    name: string;
    utype: string;
    email: string;
    phone: string;
    username: string;
    bio: string;
    password: string;
    is_verified: boolean;
    extrameta: string;
    createdAt: string;
    owner_user_id: number;
    owner_space_id: number;
    msg_read_head: number;
    disabled: boolean;
    is_deleted: boolean;
}


// /z/api/core/user/

export const getUsers = async () => {
    return iaxios.get<User[]>("/core/user");
}
