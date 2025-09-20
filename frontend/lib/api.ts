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

// User Invites
export interface UserInvite {
    id: number;
    email: string;
    role: string;
    status: string;
    invited_by: number;
    invited_as_type: string;
    expires_on: string;
    created_at: string;
}

export const getUserInvites = async () => {
    return iaxios.get<UserInvite[]>("/core/user/invites");
}

export const getUserInvite = async (id: number) => {
    return iaxios.get<UserInvite>(`/core/user/invites/${id}`);
}

export const createUserInvite = async (data: {
    email: string;
    role: string;
    invited_as_type: string;
}) => {
    return iaxios.post<UserInvite>("/core/user/invites", data);
}

export const updateUserInvite = async (id: number, data: any) => {
    return iaxios.put(`/core/user/invites/${id}`, data);
}

export const deleteUserInvite = async (id: number) => {
    return iaxios.delete(`/core/user/invites/${id}`);
}

export const resendUserInvite = async (id: number) => {
    return iaxios.post<UserInvite>(`/core/user/invites/${id}/resend`);
}

// Create User Directly
export const createUserDirectly = async (data: {
    name: string;
    email: string;
    username: string;
    utype: string;
}) => {
    return iaxios.post<User>("/core/user/create", data);
}


export interface AdminPortalData {
    popular_keywords: string[] 
    favorite_projects: any[]    
}


export const getAdminPortalData = async (portal_type: string) => {
    return iaxios.get<AdminPortalData>(`/core/self/portalData/${portal_type}`);
}


export const installPackage = async (url: string) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install`, { url });
}

export const installPackageZip = async (zip: ArrayBuffer) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install/zip`, zip, {
        headers: {
            "Content-Type": "application/zip",
        },
    });
}

export const installPackageEmbed = async (name: string) => {
    return iaxios.post<{ package_id: number }>(`/core/package/install/embed`, { name });
}

export const deletePackage = async (id: number) => {
    return iaxios.delete<void>(`/core/package/${id}`);
}

export interface EPackage {
    name: string;
    description: string;
    slug: string;
    type: string;
    tags: string;
    version: string;
}

export const listEPackages = async () => {
    return iaxios.get<EPackage[]>(`/core/package/list`);
}


export interface Package {
    id: number;
    name: string;
    description: string;
    info: string;
    slug: string;
    type: string;
    tags: string;
    version: string;
}

export interface Space {
    id: number;
    name: string;
    namespace_key: string;
    owns_namespace: boolean;
    package_id: number;
    executor_type: string;
    sub_type: string;
    owned_by: number;
    extrameta: string;
    is_initilized: boolean;
    is_public: boolean;
}

export interface InstalledSpace {
    spaces: Space[]
    packages: Package[]
}

export const listInstalledSpaces = async () => {
    return iaxios.get<InstalledSpace>(`/core/space/installed`);
}


export interface SpaceInfo {
    id: number;
    namespace_key: string;
    owns_namespace: boolean;
    package_name: string;
    package_info: string;
}

export const getSpaceInfo = async (space_key: string) => {
    return iaxios.get<SpaceInfo>(`/core/engine/space_info/${space_key}`);
}


export const authorizeSpace = async (space_key: string, space_id: number) => {
    return iaxios.post<{ token: string }>(`/core/space/authorize/${space_key}`, { space_id });
}